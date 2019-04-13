package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dyna "github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/docshelf/docshelf"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// GetPolicy fetches an existing docshelf Policy from dynamo.
func (s Store) GetPolicy(ctx context.Context, id string) (docshelf.Policy, error) {
	var policy docshelf.Policy

	if err := s.getItem(ctx, s.policyTable, "id", id, &policy); err != nil {
		return policy, errors.Wrap(err, "failed to fetch policy from dynamo")
	}

	return policy, nil
}

// PutPolicy creates a new docshelf Policy or updates an existing one in dynamo.
func (s Store) PutPolicy(ctx context.Context, policy docshelf.Policy) (string, error) {
	if policy.ID == "" {
		policy.ID = xid.New().String()
		policy.CreatedAt = time.Now()
	}

	policy.UpdatedAt = time.Now()

	marshaled, err := dyna.MarshalMap(&policy)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal policy for dynamo")
	}

	input := dynamodb.PutItemInput{
		TableName: aws.String(s.policyTable),
		Item:      marshaled,
	}

	if _, err := s.client.PutItemRequest(&input).Send(); err != nil {
		return "", errors.Wrap(err, "failed to put policy into dynamo")
	}

	return policy.ID, nil
}

// RemovePolicy deletes a policy from dynamo.
func (s Store) RemovePolicy(ctx context.Context, id string) error {
	key, err := makeKey("id", id)
	if err != nil {
		return errors.Wrap(err, "failed to make key")
	}

	input := dynamodb.DeleteItemInput{
		TableName: aws.String(s.policyTable),
		Key:       key,
	}

	if _, err := s.client.DeleteItemRequest(&input).Send(); err != nil {
		return errors.Wrap(err, "failed to remove policy from dynamo")
	}

	return nil
}
