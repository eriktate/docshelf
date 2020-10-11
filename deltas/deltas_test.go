package deltas_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/docshelf/docshelf/deltas"
)

func Test_RenderMarkdown(t *testing.T) {
	// SETUP
	input := []byte(`
	{
		"ops": [
			{
				"insert": "Hello there you ",
				"attributes": {
					"bold": true
				}
			},
			{
				"insert": "beautiful ",
				"attributes": {
					"bold": true,
					"italic": true
				}
			},
			{
				"insert": "people!",
				"attributes": {
					"bold": true
				}
			},
			{
				"insert": "Sorry, I'll stop shouting...\n"
			},
			{
				"insert": "Bullet item 1"
			},
			{
				"insert": "\n",
				"attributes": {
					"list": "bullet"
				}
			},
			{
				"insert": "Bullet item 2"
			},
			{
				"insert": "\n",
				"attributes": {
					"list": "bullet"
				}
			}
		]
	}
	`)

	var delta deltas.Delta
	if err := json.Unmarshal(input, &delta); err != nil {
		t.Fatalf("failed to unmarshal test data: %s", err)
	}

	// RUN
	md, err := delta.RenderMarkdown()
	if err != nil {
		t.Fatal("failed to render markdown")
	}

	log.Printf("Result: %s", md)
}
