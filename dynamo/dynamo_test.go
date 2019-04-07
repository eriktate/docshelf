package dynamo

import (
	"context"
	"os"
	"testing"

	"github.com/eriktate/docshelf"
)

func init() {
	os.Setenv("DS_DYNAMO_USER_TABLE", "ds_test_user")
}

func Test_New(t *testing.T) {
	// RUN
	_, err := New()

	// ASSERT
	if err != nil {
		t.Fatal(err)
	}
}

func Test_PutGetRemoveUser(t *testing.T) {
	// SETUP
	ctx := context.Background()

	store, err := New()
	if err != nil {
		t.Fatal(err)
	}

	user := docshelf.User{
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
	if err != nil {
		if _, ok := err.(docshelf.ErrRemoved); !ok {
			t.Fatal(err)
		}
	}

	// ASSERT
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

	store, err := New()
	if err != nil {
		t.Fatal(err)
	}

	user1 := docshelf.User{
		Email: "test@test.com",
		Name:  "test",
		Token: "abc123",
	}

	user2 := user1
	user2.Email = "test2@test.com"
	user3 := user1
	user3.Email = "test3@test.com"

	existing, err := store.ListUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}

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

	if len(users) != len(existing)+3 {
		t.Fatal("returned wrong number of users")
	}
}
