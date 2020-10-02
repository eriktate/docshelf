package deltas

import "reflect"

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
	ListTypeOrdered   = ListType("ordered")
	ListTypeUnordered = ListType("unordered")
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

// RenderMarkdown renders the Delta as a markdown document.
func (d Delta) RenderMarkdown() (string, error) {
	return "NOT IMPLEMENTED", nil
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
