package docshelf

import (
	"context"
	"time"
)

// A User is the identity of anyone using docshelf.
type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Token     string     `json:"token"`
	Groups    []string   `json:"groups"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

// A Group is collection of Users for the purposes of granting access.
type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Users     []string  `json:"users"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// A Doc is a full docshelf document. This includes metadata as well as content.
type Doc struct {
	Path      string    `json:"path"`
	Title     string    `json:"title"`
	IsDir     bool      `json:"isDir"`
	Content   string    `json:"content,omitempty"`
	Policy    *Policy   `json:"policy"`
	Tags      []string  `json:"tags"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// An Policy defines the users and groups that have access to a particular file path.
type Policy struct {
	ID        string    `json:"id"`
	Users     []string  `json:"users"`
	Groups    []string  `json:"groups"`
	ReadOnly  bool      `json:"readOnly"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// A DocStore knows how to store and retrieve docshelf documents.
type DocStore interface {
	GetDoc(ctx context.Context, path string) (Doc, error)
	ListDocs(ctx context.Context, path string, tags ...string) ([]Doc, error)
	PutDoc(ctx context.Context, doc Doc) error
	TagDoc(ctx context.Context, path string, tags ...string) error
	RemoveDoc(ctx context.Context, path string) error
}

// A UserStore knows how to store and retrieve docshelf users.
type UserStore interface {
	GetUser(ctx context.Context, id string) (User, error)
	// GetEmail(ctx context.Context, email string) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	PutUser(ctx context.Context, user User) (string, error)
	RemoveUser(ctx context.Context, id string) error
}

// A GroupStore knows how to store and retrieve docshelf Groups.
type GroupStore interface {
	GetGroup(ctx context.Context, id string) (Group, error)
	PutGroup(ctx context.Context, group Group) (string, error)
	RemoveGroup(ctx context.Context, id string) error
}

// A PolicyStore knows how to store and retrieve docshelf Policies.
type PolicyStore interface {
	GetPolicy(ctx context.Context, id string) (Policy, error)
	PutPolicy(ctx context.Context, policy Policy) (string, error)
	RemovePolicy(ctx context.Context, id string) (string, error)
}

// An Authenticator knows how to authenticate user credentials.
type Authenticator interface {
	Authenticate(ctx context.Context, email, token string) error
}

// A Backend is an aggregation of almost all docshelf store interfaces.
type Backend interface {
	DocStore
	UserStore
	GroupStore
	// PolicyStore
}

// A FileStore knows how to store and retrieve docshelf document contents.
type FileStore interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	RemoveFile(path string) error
	ListDir(path string) ([]string, error)
}

// An TextIndex knows how to index and search docshelf documents.
type TextIndex interface {
	Index(ctx context.Context, doc Doc) error
	Search(ctx context.Context, term string) ([]string, error)
}

// ContentString returns a Doc's content as a string.
func (d Doc) ContentString() string {
	return string(d.Content)
}
