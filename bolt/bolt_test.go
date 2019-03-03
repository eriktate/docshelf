package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/eriktate/skribe"
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
