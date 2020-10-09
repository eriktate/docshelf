package deltas

import (
	"bytes"
	"reflect"
	"strings"
)

// A Delta is a list of operations to perform on a document.
type Delta struct {
	Ops []Op `json:"ops"`
}

// An Op describes a text transformation. Can either insert text, delete text, or retain text. Additional attributes can change
// the behavior of these operations when outputing a final document.
type Op struct {
	Insert string     `json:"insert,omitempty"`     // string to to be inserted
	Delete int        `json:"delete,omitempty"`     // number of chars to delete from current position
	Retain int        `json:"retain,omitempty"`     // number of chars to retain (keep/skip) from the current position
	Attrs  Attributes `json:"attributes,omitempty"` // attributes to modify the resulting document
}

// A ListType defines what values are valid for a list attribute.
type ListType string

// ListType enum values
const (
	ListTypeOrdered = ListType("ordered")
	ListTypeBullet  = ListType("bullet")
)

// Attributes define special formatting or configuration to use when rendering a document.
type Attributes struct {
	Header    int      `json:"header,omitempty"`
	Bold      bool     `json:"bold,omitempty"`
	Italic    bool     `json:"italic,omitempty"`
	Underline bool     `json:"underline,omitempty"`
	Color     string   `json:"color,omitempty"`
	Link      string   `json:"link,omitempty"`
	List      ListType `json:"list,omitempty"`
}

// HasAttributes returns whether or not an Op has non-zero attributes defined.
func (o Op) HasAttributes() bool {
	v := reflect.ValueOf(o.Attrs)
	return !v.IsZero()
}

// IsInsert returns true if the Op is an insert.
func (o Op) IsInsert() bool {
	return o.Insert != ""
}

// IsDelete returns true if the Op is a delete.
func (o Op) IsDelete() bool {
	return o.Delete != 0
}

// IsRetain returns true if the Op is a retain.
func (o Op) IsRetain() bool {
	return o.Retain != 0
}

func resolveAttrsMD(existing, target Attributes) string {
	// To handle when to open/vs close text modifiers, we just
	// need to check if the target differs from the current state.
	// Either way, we need to add the appropriate markdown.
	mod := ""
	if target.Bold != existing.Bold {
		mod += "**"
	}

	// if the target state is italic and we currently aren't italic, add *
	// this will compound with bold if both are defined
	if target.Italic != existing.Italic {
		mod += "*"
	}

	// if the target state is underlined and we currently aren't underlined, add _
	if target.Underline != existing.Underline {
		mod += "_"
	}

	return mod
}

func trailingWS(input string) bool {
	return input[len(input)-1] == ' '
}

// func trailingNL(input string) bool {
// 	return input[len(input)-1] == '\n'
// }

// RenderMarkdown renders the Delta as a markdown document.
func (d Delta) RenderMarkdown() (string, error) {
	docBuf := bytes.NewBuffer(make([]byte, 0, 1024*5))

	var currentAttrs Attributes
	lineBuf := bytes.NewBuffer(make([]byte, 0, 1024))
	for _, op := range d.Ops {
		// need to be careful with trailing spaces because markdown doesn't
		// like them when rendering special formatting
		lineBuf.WriteString(resolveAttrsMD(currentAttrs, op.Attrs))
		lineBuf.WriteString(op.Insert)
		currentAttrs = op.Attrs

		if strings.HasSuffix(op.Insert, "\n") {
			// handle line formatting
			if op.Attrs.List == ListTypeBullet {
				docBuf.WriteString("- ")
			}

			docBuf.WriteString(lineBuf.String())
			lineBuf.Reset()
		}
	}

	return docBuf.String(), nil
}

// RenderHTML renders the Delta as an HTML document.
func (d Delta) RenderHTML() (string, error) {
	return "NOT IMPLEMENTED", nil
}

// ParseMarkdown generates a Delta from a markdown document.
func ParseMarkdown(raw string) (Delta, error) {
	return Delta{}, nil
}

// ParseHTML generates a Delta from an HTML document.
func ParseHTML(raw string) (Delta, error) {
	return Delta{}, nil
}
