package deltas

import "reflect"

// A Delta is a list of operations to perform on a document.
type Delta struct {
	Ops []Op `json:"ops"`
}

// An Operation describes a text transformation. Can either insert text, delete text, or retain text. Additional attributes can change
// the behavior of these operations when outputing a final document.
type Op struct {
	Insert string     `json:"insert,omitempty"`     // string to to be inserted
	Delete int        `json:"delete,omitempty"`     // number of chars to delete from current position
	Retain int        `json:"retain,omitempty"`     // number of chars to retain (keep/skip) from the current position
	Attrs  Attributes `json:"attributes,omitempty"` // attributes to modify the resulting document
}

type ListType string

const (
	ListTypeOrdered   = ListType("ordered")
	ListTypeUnordered = ListType("unordered")
)

type Attributes struct {
	Header    int      `json:"header,omitempty"`
	Bold      bool     `json:"bold,omitempty"`
	Italic    bool     `json:"italic,omitempty"`
	Underline bool     `json:"underline,omitempty"`
	Color     string   `json:"color,omitempty"`
	Link      string   `json:"link,omitempty"`
	List      ListType `json:"list,omitempty"`
}

func (o Op) HasAttributes() bool {
	v := reflect.ValueOf(o.Attrs)
	return !v.IsZero()
}

func (o Op) IsInsert() bool {
	return o.Insert != ""
}

func (o Op) IsDelete() bool {
	return o.Delete != 0
}

func (o Op) IsRetain() bool {
	return o.Retain != 0
}

func (d Delta) RenderMarkdown() (string, error) {
	return "NOT IMPLEMENTED", nil
}

func (d Delta) RenderHTML() (string, error) {
	return "NOT IMPLEMENTED", nil
}

func ParseMarkdown(raw string) (Delta, error) {
	return Delta{}, nil
}

func ParseHTML(raw string) (Delta, error) {
	return Delta{}, nil
}
