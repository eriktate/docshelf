package disk

import (
	"testing"
)

func Test_FileLifecycle(t *testing.T) {
	// SETUP
	root := "./documents"
	store := New(root)
	testFile := []byte("This is some test content to store!")
	testPath := "test.md"

	// RUN
	if err := store.WriteFile(testPath, testFile); err != nil {
		t.Fatal(err)
	}

	content, err := store.ReadFile(testPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveFile(testPath); err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if string(testFile) != string(content) {
		t.Fatal("Source content does not match retrieved content")
	}
}

func Test_WriteTree(t *testing.T) {
	// SETUP
	root := "./documents"
	store := New(root)
	testFile := []byte("This is some test content to store!")
	testPath := "test/test.md"

	// RUN
	if err := store.WriteFile(testPath, testFile); err != nil {
		t.Fatal(err)
	}

	content, err := store.ReadFile(testPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveFile(testPath); err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if string(testFile) != string(content) {
		t.Fatal("Source content does not match retrieved content")
	}
}

func Test_ListDir(t *testing.T) {
	// SETUP
	root := "./documents"
	store := New(root)
	testFile := []byte("This is some test content to store!")
	testPath1 := "test/test1.md"
	testPath2 := "test/test2.md"

	// RUN
	if err := store.WriteFile(testPath1, testFile); err != nil {
		t.Fatal(err)
	}

	if err := store.WriteFile(testPath2, testFile); err != nil {
		t.Fatal(err)
	}

	list, err := store.ListDir("test")

	if err := store.RemoveFile(testPath1); err != nil {
		t.Fatalf("failed to cleanup: %s", err)
	}

	if err := store.RemoveFile(testPath2); err != nil {
		t.Fatalf("failed to cleanup: %s", err)
	}

	// ASSERT
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatal("failed to list all files")
	}
}
