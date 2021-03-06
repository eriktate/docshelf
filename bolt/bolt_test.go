package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/mock"
	"github.com/rs/xid"
)

const dbName = "test.db"

func Test_New(t *testing.T) {
	// SETUP
	defer os.Remove(dbName) // cleanup database after test

	// RUN
	store, err := New(dbName, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// ASSERT
	if err != nil {
		t.Fatal(err)
	}
}

func Test_PutGetRemoveUser(t *testing.T) {
	// SETUP
	ctx := context.Background()
	defer os.Remove(dbName) // cleanup database after test

	store, err := New(dbName, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

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
	defer os.Remove(dbName) // cleanup database after test

	store, err := New(dbName, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	user1 := docshelf.User{
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
	defer os.Remove(dbName) // cleanup database after test

	store, err := New(dbName, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	group := docshelf.Group{
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
	defer os.Remove(dbName) // cleanup database after test

	store, err := New(dbName, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	policy := docshelf.Policy{
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

func Test_PutGetRemoveDoc(t *testing.T) {
	// SETUP
	ctx := context.Background()
	defer os.Remove(dbName) // cleanup database after test

	store, err := New(dbName, mock.NewFileStore(), mock.NewTextIndex(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	doc := docshelf.Doc{
		Path:      "test.md",
		Title:     "Test Document",
		Content:   "This is a test document, for testing purposes only",
		CreatedBy: xid.New().String(),
		UpdatedBy: xid.New().String(),
	}

	// RUN
	if _, err := store.PutDoc(ctx, doc); err != nil {
		t.Fatal(err)
	}

	getDoc, err := store.GetDoc(ctx, doc.Path)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveDoc(ctx, doc.Path); err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if getDoc.Title != doc.Title {
		t.Fatal("doc titles don't match")
	}

	if string(getDoc.Content) != string(doc.Content) {
		t.Fatal("doc content doesn't match")
	}
}

// TODO (erik): This test isn't exhaustive enough. Need to fix.
func Test_ListDocs(t *testing.T) {
	// SETUP
	ctx := context.Background()

	store, err := New(dbName, mock.NewFileStore(), mock.NewTextIndex(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.Remove(dbName) // cleanup database after test

	doc1 := docshelf.Doc{
		Path:      "test1.md",
		Title:     "Test Document 1",
		Content:   "This is a test document, for testing purposes only",
		CreatedBy: xid.New().String(),
		UpdatedBy: xid.New().String(),
	}

	doc2 := docshelf.Doc{
		Path:      "test2.md",
		Title:     "Test Document 2",
		Content:   "This is a test document, for testing purposes only",
		CreatedBy: xid.New().String(),
		UpdatedBy: xid.New().String(),
	}

	// RUN
	if _, err := store.PutDoc(ctx, doc1); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutDoc(ctx, doc2); err != nil {
		t.Fatal(err)
	}

	list, err := store.ListDocs(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if len(list) != 2 {
		t.Fatal("listing didn't return enough results")
	}
}

func Test_TagLifecycle(t *testing.T) {
	// SETUP
	ctx := context.Background()

	store, err := New(dbName, mock.NewFileStore(), mock.NewTextIndex(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	defer os.Remove(dbName) // cleanup database after test

	doc1 := docshelf.Doc{
		Path:      "test1.md",
		Title:     "Test Document 1",
		Content:   "This is a test document, for testing purposes only",
		CreatedBy: xid.New().String(),
		UpdatedBy: xid.New().String(),
	}

	doc2 := docshelf.Doc{
		Path:      "test2.md",
		Title:     "Test Document 2",
		Content:   "This is a test document, for testing purposes only",
		CreatedBy: xid.New().String(),
		UpdatedBy: xid.New().String(),
	}

	// RUN
	if _, err := store.PutDoc(ctx, doc1); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutDoc(ctx, doc2); err != nil {
		t.Fatal(err)
	}

	if err := store.TagDoc(ctx, doc1.Path, "test", "one"); err != nil {
		t.Fatal(err)
	}

	if err := store.TagDoc(ctx, doc2.Path, "test", "two"); err != nil {
		t.Fatal(err)
	}

	testTag, err := store.ListDocs(ctx, "", "test")
	if err != nil {
		t.Fatal(err)
	}

	oneTag, err := store.ListDocs(ctx, "", "one")
	if err != nil {
		t.Fatal(err)
	}

	twoTag, err := store.ListDocs(ctx, "", "two")
	if err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if len(testTag) != 2 {
		t.Fatal("listing didn't return enough results")
	}

	if len(oneTag) != 1 && oneTag[0].Path == doc1.Path {
		t.Fatal("listing returned wrong results for tag 'one'")
	}

	if len(twoTag) != 1 && twoTag[0].Path == doc2.Path {
		t.Fatal("listing returned wrong results for tag 'two'")
	}
}
