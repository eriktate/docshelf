package bleve

import (
	"context"
	"os"

	"github.com/eriktate/docshelf"

	"testing"
)

const testBlevePath = "test_docshelf.bleve"

func init() {
	if err := os.Setenv("DS_INDEX_PATH", testBlevePath); err != nil {
		panic(err)
	}
}

func Test_IndexSearch(t *testing.T) {
	defer os.RemoveAll(testBlevePath)
	// SETUP
	ctx := context.Background()
	idx, err := New()
	if err != nil {
		t.Fatal(err)
	}

	doc1 := docshelf.Doc{
		Path:    "testPath1",
		Content: []byte("This is a test document about unicorns"),
	}

	doc2 := docshelf.Doc{
		Path:    "testPath2",
		Content: []byte("This is a test document about gophers"),
	}

	// RUN
	if err := idx.Index(ctx, doc1); err != nil {
		t.Fatal(err)
	}

	if err := idx.Index(ctx, doc2); err != nil {
		t.Fatal(err)
	}

	unicornResults, err := idx.Search(ctx, "unicorn")
	if err != nil {
		t.Fatal(err)
	}

	gopherResults, err := idx.Search(ctx, "gopher")
	if err != nil {
		t.Fatal(err)
	}

	documentResults, err := idx.Search(ctx, "document")
	if err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if len(unicornResults) == 0 || unicornResults[0] != doc1.Path {
		t.Fatal("search returned incorrent document")
	}

	if len(gopherResults) == 0 || gopherResults[0] != doc2.Path {
		t.Fatal("search returned incorrent document")
	}

	if len(documentResults) != 2 {
		t.Fatal("search returned incorrect documents")
	}
}
