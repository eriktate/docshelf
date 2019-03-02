package disk

import "testing"

func Test_RWDFile(t *testing.T) {
	// SETUP
	root := "./documents"
	store := New(root)
	testFile := []byte("This is some test content to store!")
	testName := "test.md"

	// RUN
	if err := store.WriteFile(testName, testFile); err != nil {
		t.Fatal(err)
	}

	content, err := store.ReadFile(testName)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.RemoveFile(testName); err != nil {
		t.Fatal(err)
	}

	// ASSERT
	if string(testFile) != string(content) {
		t.Fatal("Source content does not match retrieved content")
	}
}
