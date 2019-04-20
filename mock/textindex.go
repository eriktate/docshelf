package mock

import (
	"context"

	"github.com/docshelf/docshelf"
)

// TextIndex is a mock implementation of the docshelf.TextIndex interface.
type TextIndex struct {
	SearchFn     func(ctx context.Context, term string) ([]string, error)
	SearchCalled int

	IndexFn     func(ctx context.Context, doc docshelf.Doc) error
	IndexCalled int

	Err error
}

// NewTextIndex returns a new mock TextIndex struct.
func NewTextIndex(err error) *TextIndex {
	return &TextIndex{
		Err: err,
	}
}

// Search mocks the docshelf.TextIndex interface.
func (m *TextIndex) Search(ctx context.Context, term string) ([]string, error) {
	m.SearchCalled++
	if m.Err != nil {
		return nil, m.Err
	}

	if m.SearchFn != nil {
		return m.SearchFn(ctx, term)
	}

	return nil, nil
}

// Index mocks the docshelf.TextIndex interface.
func (m *TextIndex) Index(ctx context.Context, doc docshelf.Doc) error {
	m.IndexCalled++
	if m.Err != nil {
		return m.Err
	}

	if m.IndexFn != nil {
		return m.IndexFn(ctx, doc)
	}

	return nil
}
