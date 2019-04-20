package bleve

import (
	"context"
	"errors"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/docshelf/docshelf"
	"github.com/docshelf/docshelf/env"
)

const defIndexPath = "docshelf.bleve"

// An Index implements the docshelf.TextIndex interface.
type Index struct {
	idx bleve.Index
}

// New returns a new bleve Index.
func New() (Index, error) {
	path := env.GetEnvString("DS_INDEX_PATH", defIndexPath)
	stat, err := os.Stat(path)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		idx, err := bleve.New(env.GetEnvString("DS_INDEX_PATH", defIndexPath), mapping)
		if err != nil {
			return Index{}, err
		}

		return Index{idx}, nil
	}

	if !stat.IsDir() {
		return Index{}, errors.New("bleve index path exists, but isn't a folder")
	}

	idx, err := bleve.Open(path)
	if err != nil {
		return Index{}, err
	}

	return Index{idx}, nil
}

// Index takes a docshelf Doc and indexes it in bleve.
func (i Index) Index(ctx context.Context, doc docshelf.Doc) error {
	return i.idx.Index(doc.Path, doc.ContentString())
}

// Search takes a search term and returns all doc paths that match.
func (i Index) Search(ctx context.Context, query string) ([]string, error) {
	req := bleve.NewSearchRequest(bleve.NewFuzzyQuery(query))
	res, err := i.idx.Search(req)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(res.Hits))
	for i, hit := range res.Hits {
		ids[i] = hit.ID
	}

	return ids, nil
}
