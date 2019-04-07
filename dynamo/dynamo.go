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
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

const (
	defUserTable = "ds_user"
	defDocTable  = "ds_doc"
	defTagTable  = "ds_tag"
)

// A Store has methods that know how to interact with docshelf data in Dynamo.
type Store struct {
	client dynamodbiface.DynamoDBAPI
	fs     docshelf.FileStore

	userTable string
	docTable  string
	tagTable  string
}

// New creates a new Store struct.
func New(fs docshelf.FileStore) (Store, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return Store{}, err
	}

	cfg.Region = endpoints.UsEast1RegionID
	svc := dynamodb.New(cfg)

	store := Store{
		client:    svc,
		fs:        fs,
		userTable: env.GetEnvString("DS_DYNAMO_USER_TABLE", defUserTable),
		docTable:  env.GetEnvString("DS_DYNAMO_DOC_TABLE", defDocTable),
		tagTable:  env.GetEnvString("DS_DYNAMO_TAG_TABLE", defTagTable),
	}

	return store, store.ensureTables()
}

// GetUser fetches an existing docshelf User from dynamodb.
func (s Store) GetUser(ctx context.Context, id string) (docshelf.User, error) {
	var user docshelf.User

	if err := s.getItem(ctx, s.userTable, "id", id, &user); err != nil {
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

	// TODO (erik): Figure out if this actually works with the GSI.
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

// GetDoc fetches a docshelf Document from dynamodb. It will also read and package the Content
// form an underlying FileStore.
func (s Store) GetDoc(ctx context.Context, path string) (docshelf.Doc, error) {
	var doc docshelf.Doc

	if err := s.getItem(ctx, s.docTable, "path", path, &doc); err != nil {
		return doc, err
	}

	content, err := s.fs.ReadFile(path)
	if err != nil {
		return doc, errors.Wrap(err, "failed to read file from file store")
	}

	doc.Content = content
	return doc, nil
}

// ListDocs fetches a slice of docshelf Document metadata from dynamodb that fit the give prefix
// and tags supplied.
func (s Store) ListDocs(ctx context.Context, prefix string, tags ...string) ([]docshelf.Doc, error) {
	// prefer to filter by tag first if supplied.
	if len(tags) > 0 {
		tagged, err := s.listTaggedDocs(ctx, tags)
		if err != nil {
			return nil, err
		}

		// short circuit if no prefix supplied
		if prefix == "" {
			return tagged, nil
		}

		listing := make([]docshelf.Doc, 0, len(tagged))
		for _, doc := range tagged {
			if strings.HasPrefix(doc.Path, prefix) {
				listing = append(listing, doc)
			}
		}

		return listing, nil
	}

	input := dynamodb.ScanInput{
		TableName: aws.String(s.docTable),
	}

	res, err := s.client.ScanRequest(&input).Send()
	if err != nil {
		return nil, err
	}

	var docs []docshelf.Doc
	if err := dyna.UnmarshalListOfMaps(res.Items, &docs); err != nil {
		return nil, err
	}

	// TODO (erik): Duplicate of code above. Should refactor to combine.
	listing := make([]docshelf.Doc, 0, len(docs))
	for _, doc := range docs {
		if strings.HasPrefix(doc.Path, prefix) {
			listing = append(listing, doc)
		}
	}

	return listing, nil
}

func (s Store) listTaggedDocs(ctx context.Context, tags []string) ([]docshelf.Doc, error) {
	var paths []string
	for _, t := range tags {
		var tag Tag
		if err := s.getItem(ctx, s.tagTable, "tag", t, &tag); err != nil {
			return nil, err
		}

		if paths == nil {
			paths = tag.Paths
		} else {
			paths = intersect(paths, tag.Paths)
		}
	}

	var docs []docshelf.Doc
	for _, path := range paths {
		var doc docshelf.Doc
		if err := s.getItem(ctx, s.docTable, "path", path, &doc); err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// PutDoc creates or updates an existing docshelf Doc in dynamodb. It will also store the
// Content in an underlying FileStore.
func (s Store) PutDoc(ctx context.Context, doc docshelf.Doc) error {
	if doc.Path == "" {
		return errors.New("can not create a new doc without a path")
	}

	if _, err := s.GetDoc(ctx, doc.Path); err != nil {
		if !docshelf.CheckDoesNotExist(err) {
			return errors.Wrap(err, "could not verify existing file")
		}

		doc.CreatedAt = time.Now()
	}

	doc.UpdatedAt = time.Now()

	if err := s.fs.WriteFile(doc.Path, doc.Content); err != nil {
		return errors.Wrap(err, "failed to write doc to file store")
	}

	doc.Content = nil // need to clear content before storing doc

	marshaled, err := dyna.MarshalMap(&doc)
	if err != nil {
		return errors.Wrap(err, "failed to marshal doc for dynamo")
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(s.docTable),
		Item:      marshaled,
	}

	if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
		if err := s.fs.RemoveFile(doc.Path); err != nil { // need to rollback file storage if doc failes
			return errors.Wrapf(err, "cleanup failed for file: %s", doc.Path)
		}

		return errors.Wrap(err, "failed to put doc into dynamo")
	}

	return nil
}

// A Tag represents the data structure of a tag specific to dynamo.
type Tag struct {
	Tag   string   `json:"tag"`
	Paths []string `json:"paths"`
}

// TagDoc tags an existing document with the given tags.
// TODO (erik): This is a mirror of the bolt implementation. Need to research and find out
// if there's a more efficient way to get this behavior out of dynamo.
func (s Store) TagDoc(ctx context.Context, path string, tags ...string) error {
	for _, t := range tags {
		var tag Tag
		if err := s.getItem(ctx, s.tagTable, "tag", t, &tag); err != nil {
			return err
		}

		// short circuit if the tag alrady contains the path or no tag was returned.
		if contains(tag.Paths, path) {
			continue
		}

		if tag.Tag == "" {
			tag.Tag = t
		}

		tag.Paths = append(tag.Paths, path)
		marshaled, err := dyna.MarshalMap(&tag)
		if err != nil {
			return err
		}

		input := dynamodb.PutItemInput{
			TableName: aws.String(s.tagTable),
			Item:      marshaled,
		}

		if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
			return err
		}
	}

	return nil
}

// RemoveDoc removes a docshelf Doc from dynamo as well as the underlying FileStore.
func (s Store) RemoveDoc(ctx context.Context, path string) error {
	if err := s.fs.RemoveFile(path); err != nil {
		return errors.Wrap(err, "failed to remove doc from file store")
	}

	key, err := makeKey("path", path)
	if err != nil {
		return errors.Wrap(err, "failed to make key")
	}

	input := dynamodb.DeleteItemInput{
		TableName: aws.String(s.docTable),
		Key:       key,
	}

	if _, err := s.client.DeleteItemRequest(&input).Send(); err != nil {
		return errors.Wrap(err, "failed to delete doc from dynamo")
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
		if err := ensureTable(s.client, s.userTable, userTableInput(s.userTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ensureTable(s.client, s.docTable, docTableInput(s.docTable)); err != nil {
			ensureErr = err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ensureTable(s.client, s.tagTable, tagTableInput(s.tagTable)); err != nil {
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

func userTableInput(userTable string) dynamodb.CreateTableInput {
	gsiKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("email"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	gsi := dynamodb.GlobalSecondaryIndex{
		IndexName:  aws.String(fmt.Sprintf("%s_email_idx", userTable)),
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
		TableName:              aws.String(userTable),
		BillingMode:            dynamodb.BillingModePayPerRequest,
		AttributeDefinitions:   attrDef,
		GlobalSecondaryIndexes: []dynamodb.GlobalSecondaryIndex{gsi},
		KeySchema:              []dynamodb.KeySchemaElement{hashKey},
	}
}

func docTableInput(docTable string) dynamodb.CreateTableInput {
	hashKey := dynamodb.KeySchemaElement{
		AttributeName: aws.String("path"),
		KeyType:       dynamodb.KeyTypeHash,
	}

	attrDef := []dynamodb.AttributeDefinition{
		makeAttrDef("path", dynamodb.ScalarAttributeTypeS),
	}

	return dynamodb.CreateTableInput{
		TableName:            aws.String(docTable),
		BillingMode:          dynamodb.BillingModePayPerRequest,
		AttributeDefinitions: attrDef,
		KeySchema:            []dynamodb.KeySchemaElement{hashKey},
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
		return err
	}

	if err := dyna.UnmarshalMap(res.Item, out); err != nil {
		return err
	}

	return nil
}
