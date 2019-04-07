package dynamo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbiface"
	"github.com/eriktate/docshelf"
	"github.com/eriktate/docshelf/env"
	"github.com/rs/xid"
)

const (
	userTable = "ds_user"
	docTable  = "ds_doc"
)

func (s Store) userTableInput() dynamodb.CreateTableInput {
	gsiKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("email"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	gsi := dynamodb.GlobalSecondaryIndex{
		IndexName:  aws.String(fmt.Sprintf("%s_email_idx", s.userTable)),
		KeySchema:  []dynamodb.KeySchemaElement{gsiKey},
		Projection: &dynamodb.Projection{ProjectionType: dynamodb.ProjectionTypeKeysOnly},
	}

	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("id"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("id", dynamodb.ScalarAttributeTypeS),
		makeAttrDef("email", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:              aws.String(s.userTable),
		BillingMode:            dynamodb.BillingModePayPerRequest,
		AttributeDefinitions:   attrDef,
		GlobalSecondaryIndexes: []dynamodb.GlobalSecondaryIndex{gsi},
		KeySchema:              []dynamodb.KeySchemaElement{hashKey},
	}
}

// A Store has methods that know how to interact with docshelf data in Dynamo.
type Store struct {
	client dynamodbiface.DynamoDBAPI

	userTable string
	docTable  string
}

// New creates a new Store struct.
func New() (Store, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return Store{}, err
	}

	cfg.Region = endpoints.UsEast1RegionID
	svc := dynamodb.New(cfg)

	store := Store{
		client:    svc,
		userTable: env.GetEnvString("DS_DYNAMO_USER_TABLE", userTable),
	}

	return store, store.ensureTables()
}

// GetUser fetches an existing docshelf User from dynamodb.
func (s Store) GetUser(ctx context.Context, id string) (docshelf.User, error) {
	var user docshelf.User
	key, err := makeKey("id", id)
	if err != nil {
		return user, err
	}

	input := dynamodb.GetItemInput{
		TableName: aws.String(s.userTable),
		Key:       key,
	}

	res, err := s.client.GetItemRequest(&input).Send()
	if err != nil {
		return user, err
	}

	if err := dyna.UnmarshalMap(res.Item, &user); err != nil {
		return user, err
	}

	return user, nil
}

// GetEmail fetches an existing docshelf User from dynamodb given an email.
func (s Store) GetEmail(ctx context.Context, email string) (docshelf.User, error) {
	var user docshelf.User
	key, err := makeKey("email", email)
	if err != nil {
		return user, err
	}

	input := dynamodb.GetItemInput{
		TableName: aws.String(s.userTable),
		Key:       key,
	}

	res, err := s.client.GetItemRequest(&input).Send()
	if err != nil {
		return user, err
	}

	if err := dyna.UnmarshalMap(res.Item, &user); err != nil {
		return user, err
	}

	return user, nil
}

// ListUsers returns all docshelf Users stored in dynamodb.
func (s Store) ListUsers(ctx context.Context) ([]docshelf.User, error) {
	input := dynamodb.ScanInput{
		TableName: aws.String(s.userTable),
	}

	res, err := s.client.ScanRequest(&input).Send()
	if err != nil {
		return nil, err
	}

	var users []docshelf.User
	if err := dyna.UnmarshalListOfMaps(res.Items, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// PutUser creates a new docshelf User or updates an existing one in dynamodb.
func (s Store) PutUser(ctx context.Context, user docshelf.User) (string, error) {
	if user.ID == "" {
		user.ID = xid.New().String()
		user.CreatedAt = time.Now()
	}

	user.UpdatedAt = time.Now()

	marshaled, err := dyna.MarshalMap(&user)
	if err != nil {
		return "", err
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(s.userTable),
		Item:      marshaled,
	}

	if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
		return "", err
	}

	return user.ID, nil
}

// RemoveUser marks a user as deleted in boltdb.
// TODO (erik): Right now this is pretty lazy. Should come back and write the expression
// to update only the DeletedAt field rather than overwriting the entire Item.
func (s Store) RemoveUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	user.DeletedAt = &now

	if _, err := s.PutUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func makeKey(key string, value interface{}) (map[string]dynamodb.AttributeValue, error) {
	v, err := dyna.Marshal(value)
	if err != nil {
		return nil, err
	}

	return map[string]dynamodb.AttributeValue{key: *v}, nil
}

// ensureTables concurrently ensures dynamo tables are created. Doing this in parallel
// significantly reduces the wait time for dynamo to be bootstrapped.
func (s Store) ensureTables() error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var ensureErr error
	go func() {
		defer wg.Done()
		if err := ensureTable(s.client, s.userTable, s.userTableInput()); err != nil {
			ensureErr = err
		}
	}()

	wg.Wait()
	return ensureErr
}

func ensureTable(svc dynamodbiface.DynamoDBAPI, table string, input dynamodb.CreateTableInput) error {
	describe := dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	}

	if _, err := svc.DescribeTableRequest(&describe).Send(); err != nil {
		// TODO (erik): This seems like a really fragile err check. Need to find a better way
		// to do this.
		if strings.Contains(err.Error(), dynamodb.ErrCodeResourceNotFoundException) {
			if _, err := svc.CreateTableRequest(&input).Send(); err != nil {
				return err
			}

			if err := svc.WaitUntilTableExists(&describe); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func makeAttrDef(key string, attrType dynamodb.ScalarAttributeType) dynamodb.AttributeDefinition {
	return dynamodb.AttributeDefinition{
		AttributeType: attrType,
		AttributeName: aws.String(key),
	}
}
