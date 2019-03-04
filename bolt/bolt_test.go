package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
	"github.com/rs/xid"
)

func Test_New(t *testing.T) {
	// SETUP
	dbName := "test.db"
	defer os.Remove(dbName) // cleanup database after test
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// RUN
	_, err = New(db)

	// ASSERT
	if err != nil {
		t.Fatal(err)
	}
}

func Test_PutGetRemoveUser(t *testing.T) {
	// SETUP
	ctx := context.Background()
	dbName := "test.db"
	defer os.Remove(dbName) // cleanup database after test
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	user := skribe.User{
		Email: "test@test.com",
		Name:  "test",
		Token: "abc123",
	}

	// RUN
	id, err := store.PutUser(ctx, user)
	if err != nil {
		t.Fatal(err)
	}

	getUser, err := store.GetUser(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveUser(ctx, id); err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUser(ctx, id)
	if err != nil && err != ErrUserRemoved {
		t.Fatal(err)
	}

	// ASSERT
	if err != ErrUserRemoved {
		t.Fatal("fetching removed user didn't return the correct error")
	}
	if getUser.Email != user.Email {
		t.Fatal("emails don't match")
	}

	if getUser.Name != user.Name {
		t.Fatal("names don't match")
	}

	if getUser.Token != user.Token {
		t.Fatal("tokens don't match")
	}

	if getUser.CreatedAt.IsZero() {
		t.Fatal("CreatedAt not set properly")
	}

	if getUser.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt not set properly")
	}
}

func Test_ListUsers(t *testing.T) {
	// SETUP
	ctx := context.Background()
	dbName := "test.db"
	defer os.Remove(dbName) // cleanup database after test
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	user1 := skribe.User{
		Email: "test@test.com",
		Name:  "test",
		Token: "abc123",
	}

	user2 := user1
	user2.Email = "test2@test.com"
	user3 := user1
	user3.Email = "test3@test.com"

	if _, err := store.PutUser(ctx, user1); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutUser(ctx, user2); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutUser(ctx, user3); err != nil {
		t.Fatal(err)
	}

	// RUN
	users, err := store.ListUsers(ctx)

	// ASSERT
	if err != nil {
		t.Fatal(err)
	}

	if len(users) != 3 {
		t.Fatal("returned wrong number of users")
	}
}

func Test_PutGetRemoveGroup(t *testing.T) {
	// SETUP
	ctx := context.Background()
	dbName := "test.db"
	defer os.Remove(dbName) // cleanup database after test
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	group := skribe.Group{
		Name:  "test",
		Users: []string{xid.New().String(), xid.New().String(), xid.New().String()},
	}

	// RUN
	id, err := store.PutGroup(ctx, group)
	if err != nil {
		t.Fatal(err)
	}

	getGroup, err := store.GetGroup(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveGroup(ctx, id); err != nil {
		t.Fatal(err)
	}

	if _, err := store.GetGroup(ctx, id); err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if getGroup.Name != group.Name {
		t.Fatal("group names don't match")
	}

	if len(getGroup.Users) != len(group.Users) {
		t.Fatal("group users aren't the same length")
	}

	if getGroup.CreatedAt.IsZero() {
		t.Fatal("group CreatedAt not set properly")
	}

	if getGroup.UpdatedAt.IsZero() {
		t.Fatal("group UpdatedAt not set properly")
	}
}

func Test_PutGetRemovePolicy(t *testing.T) {
	// SETUP
	ctx := context.Background()
	dbName := "test.db"
	defer os.Remove(dbName) // cleanup database after test
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	policy := skribe.Policy{
		Users:  []string{xid.New().String(), xid.New().String(), xid.New().String()},
		Groups: []string{xid.New().String(), xid.New().String(), xid.New().String()},
	}

	// RUN
	id, err := store.PutPolicy(ctx, policy)
	if err != nil {
		t.Fatal(err)
	}

	getPolicy, err := store.GetPolicy(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemovePolicy(ctx, id); err != nil {
		t.Fatal(err)
	}

	if _, err := store.GetPolicy(ctx, id); err != nil {
		t.Fatal(err)
	}

	// ASSERT

	if len(getPolicy.Users) != len(policy.Users) {
		t.Fatal("policy users aren't the same length")
	}

	if len(getPolicy.Groups) != len(policy.Groups) {
		t.Fatal("policy groups aren't the same length")
	}

	if getPolicy.CreatedAt.IsZero() {
		t.Fatal("policy CreatedAt not set properly")
	}

	if getPolicy.UpdatedAt.IsZero() {
		t.Fatal("policy UpdatedAt not set properly")
	}
}
