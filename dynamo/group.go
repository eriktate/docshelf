package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/eriktate/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetGroup fetches an existing docshelf Group from dynamodb.
func (s Store) GetGroup(ctx context.Context, id string) (docshelf.Group, error) {
	var group docshelf.Group

	if err := s.getItem(ctx, s.groupTable, "id", id, &group); err != nil {
		return group, errors.Wrap(err, "failed to fetch group from dynamo")
	}

	return group, nil
}

// PutGroup creates a new docshelf User or updates an existing one in boltdb.
func (s Store) PutGroup(ctx context.Context, group docshelf.Group) (string, error) {
	if group.ID == "" {
		group.ID = xid.New().String()
		group.CreatedAt = time.Now()
	}

	group.UpdatedAt = time.Now()

	marshaled, err := dyna.MarshalMap(&group)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal group for dynamo")
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(s.groupTable),
		Item:      marshaled,
	}

	if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
		return "", errors.Wrap(err, "failed to put group into dynamo")
	}

	return group.ID, nil
}

// RemoveGroup deletes a docshelf Group from dynamo.
func (s Store) RemoveGroup(ctx context.Context, id string) error {
	key, err := makeKey("id", id)
	if err != nil {
		return errors.Wrap(err, "failed to make key to delete from dynamo")
	}

	input := dynamodb.DeleteItemInput{
		TableName: aws.String(s.groupTable),
		Key:       key,
	}

	if _, err := s.client.DeleteItemRequest(&input).Send(); err != nil {
		return errors.Wrap(err, "failed to remove group from dynamo")
	}

	return nil
}
