package dynamo

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/expression"
	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/env"
	"github.com/sirupsen/logrus"
)

const (
	defUserTable   = "docshelf_user"
	defDocTable    = "docshelf_doc"
	defTagTable    = "docshelf_tag"
	defGroupTable  = "docshelf_group"
	defPolicyTable = "docshelf_policy"
)

// A Store has methods that know how to interact with docshelf data in Dynamo.
type Store struct {
	client dynamodbiface.DynamoDBAPI
	fs     docshelf.FileStore
	ti     docshelf.TextIndex
	log    *logrus.Logger

	userTable   string
	docTable    string
	tagTable    string
	groupTable  string
	policyTable string

	userEmailIndex string
	docIDIndex     string
}

// New creates a new Store struct.
func New(fs docshelf.FileStore, ti docshelf.TextIndex, logger *logrus.Logger) (Store, error) {
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(ioutil.Discard)
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return Store{}, err
	}

	cfg.Region = endpoints.UsEast1RegionID
	svc := dynamodb.New(cfg)

	store := Store{
		client:      svc,
		fs:          fs,
		ti:          ti,
		log:         logger,
		userTable:   env.GetEnvString("DS_DYNAMO_USER_TABLE", defUserTable),
		docTable:    env.GetEnvString("DS_DYNAMO_DOC_TABLE", defDocTable),
		tagTable:    env.GetEnvString("DS_DYNAMO_TAG_TABLE", defTagTable),
		groupTable:  env.GetEnvString("DS_DYNAMO_GROUP_TABLE", defGroupTable),
		policyTable: env.GetEnvString("DS_DYNAMO_POLICY_TABLE", defPolicyTable),
	}

	// set secondary indices
	store.userEmailIndex = fmt.Sprintf("%s_email_idx", store.userTable)
	store.docIDIndex = fmt.Sprintf("%s_id_idx", store.docTable)

	return store, store.ensureTables()
}

func makeKey(key string, value interface{}) (map[string]dynamodb.AttributeValue, error) {
	v, err := dyna.Marshal(value)
	if err != nil {
		return nil, err
	}

	return map[string]dynamodb.AttributeValue{key: *v}, nil
}

func (s Store) getItem(ctx context.Context, table, keyName, key string, out interface{}) error {
	k, err := makeKey(keyName, key)
	if err != nil {
		return err
	}

	input := dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key:       k,
	}

	res, err := s.client.GetItemRequest(&input).Send()
	if err != nil {
		if strings.Contains(err.Error(), dynamodb.ErrCodeResourceNotFoundException) {
			return docshelf.NewErrNotFound(err.Error())
		}
	}

	if err := dyna.UnmarshalMap(res.Item, out); err != nil {
		return err
	}

	return nil
}

func (s Store) getItemsGsi(ctx context.Context, table, idx, keyName, key string, out interface{}) error {
	keyCond := expression.Key(keyName).Equal(expression.Value(key))
	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return err
	}

	input := dynamodb.QueryInput{
		TableName:                 aws.String(table),
		IndexName:                 aws.String(idx),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	res, err := s.client.QueryRequest(&input).Send()
	if err != nil {
		if strings.Contains(err.Error(), dynamodb.ErrCodeResourceNotFoundException) {
			return docshelf.NewErrNotFound(err.Error())
		}

		return err
	}

	if err := dyna.UnmarshalListOfMaps(res.Items, out); err != nil {
		return err
	}

	return nil
}

// ensureTables concurrently ensures dynamo tables are created. Doing this in parallel
// significantly reduces the wait time for dynamo to be bootstrapped.
func (s Store) ensureTables() error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	var ensureErr error
	go func() {
		defer wg.Done()
		if err := s.ensureTable(s.userTable, userTableInput(s.userTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.ensureTable(s.docTable, docTableInput(s.docTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.ensureTable(s.tagTable, tagTableInput(s.tagTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.ensureTable(s.groupTable, groupTableInput(s.groupTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.ensureTable(s.policyTable, policyTableInput(s.policyTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Wait()
	return ensureErr
}

func (s Store) ensureTable(table string, input dynamodb.CreateTableInput) error {
	describe := dynamodb.DescribeTableInput{
		TableName: aws.String(table),
	}

	if _, err := s.client.DescribeTableRequest(&describe).Send(); err != nil {
		// TODO (erik): This seems like a really fragile err check. Need to find a better way
		// to do this.
		if strings.Contains(err.Error(), dynamodb.ErrCodeResourceNotFoundException) {
			s.log.WithField("table", table).Info("table not found, provisioning...")
			if _, err := s.client.CreateTableRequest(&input).Send(); err != nil {
				return err
			}

			if err := s.client.WaitUntilTableExists(&describe); err != nil {
				return err
			}

			s.log.WithField("table", table).Info("finished provisioning")
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

func userTableInput(userTable string) dynamodb.CreateTableInput {
	gsiKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("email"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	gsi := dynamodb.GlobalSecondaryIndex{
		IndexName: aws.String(fmt.Sprintf("%s_email_idx", userTable)),
		KeySchema: []dynamodb.KeySchemaElement{gsiKey},
		Projection: &dynamodb.Projection{
			ProjectionType:   dynamodb.ProjectionTypeInclude,
			NonKeyAttributes: []string{"token"},
		},
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
		TableName:              aws.String(userTable),
		BillingMode:            dynamodb.BillingModePayPerRequest,
		AttributeDefinitions:   attrDef,
		GlobalSecondaryIndexes: []dynamodb.GlobalSecondaryIndex{gsi},
		KeySchema:              []dynamodb.KeySchemaElement{hashKey},
	}
}

func docTableInput(docTable string) dynamodb.CreateTableInput {
	gsiKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("id"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	gsi := dynamodb.GlobalSecondaryIndex{
		IndexName:  aws.String(fmt.Sprintf("%s_id_idx", docTable)),
		KeySchema:  []dynamodb.KeySchemaElement{gsiKey},
		Projection: &dynamodb.Projection{ProjectionType: dynamodb.ProjectionTypeKeysOnly},
	}

	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("path"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("path", dynamodb.ScalarAttributeTypeS),
		makeAttrDef("id", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:              aws.String(docTable),
		BillingMode:            dynamodb.BillingModePayPerRequest,
		AttributeDefinitions:   attrDef,
		GlobalSecondaryIndexes: []dynamodb.GlobalSecondaryIndex{gsi},
		KeySchema:              []dynamodb.KeySchemaElement{hashKey},
	}
}

func tagTableInput(tagTable string) dynamodb.CreateTableInput {
	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("tag"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("tag", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:            aws.String(tagTable),
		BillingMode:          dynamodb.BillingModePayPerRequest,
		AttributeDefinitions: attrDef,
		KeySchema:            []dynamodb.KeySchemaElement{hashKey},
	}
}

func groupTableInput(groupTable string) dynamodb.CreateTableInput {
	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("id"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("id", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:            aws.String(groupTable),
		BillingMode:          dynamodb.BillingModePayPerRequest,
		AttributeDefinitions: attrDef,
		KeySchema:            []dynamodb.KeySchemaElement{hashKey},
	}
}

func policyTableInput(policyTable string) dynamodb.CreateTableInput {
	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("id"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("id", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:            aws.String(policyTable),
		BillingMode:          dynamodb.BillingModePayPerRequest,
		AttributeDefinitions: attrDef,
		KeySchema:            []dynamodb.KeySchemaElement{hashKey},
	}
}

// TODO (erik): Duplicated code shared with bolt backend. Should probably consolidate.
func intersect(left, right []string) []string {
	intersection := make([]string, 0)
	for _, el := range left {
		if contains(right, el) {
			intersection = append(intersection, el)
		}
	}

	return intersection
}

func contains(slice []string, el string) bool {
	for _, s := range slice {
		if s == el {
			return true
		}
	}

	return false
}
