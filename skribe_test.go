package skribe

import (
	"testing"
)

func Test_ContentString(t *testing.T) {
	// SETUP
	doc := Doc{
		Content: []byte("hello world"),
	}

	// RUN
	content := doc.ContentString()

	// ASSERT
	if content != "hello world" {
		t.Fatal("content returned incorrectly")
	}
}
