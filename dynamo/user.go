package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/docshelf/docshelf"
	"github.com/rs/xid"
)

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
