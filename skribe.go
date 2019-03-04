package skribe

import (
	"context"
	"time"
)

// A User is the identity of anyone using skribe.
type User struct {
	ID        string
	Email     string
	Name      string
	Token     string
	Groups    []string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// A Group is collection of Users for the purposes of granting access.
type Group struct {
	ID        string
	Name      string
	Users     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// A Doc is a full skribe document. This includes metadata as well as content.
type Doc struct {
	ID        string
	FilePath  string
	IsDir     bool
	Content   []byte
	Policy    *Policy
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy string
}

// An Policy defines the users and groups that have access to a particular file path.
type Policy struct {
	ID        string
	Users     []string
	Groups    []string
	ReadOnly  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// A DocStore knows how to store and retrieve skribe documents.
type DocStore interface {
	GetDoc(ctx context.Context, id string) (Doc, error)
	GetPath(ctx context.Context, path string) (Doc, error)
	ListPath(ctx context.Context, path string) ([]Doc, error)
	PutDoc(ctx context.Context, doc Doc) (string, error)
	RemoveDoc(ctx context.Context, id string) (string, error)
}

// A UserStore knows how to store and retrieve skribe users.
type UserStore interface {
	GetUser(ctx context.Context, id string) (User, error)
	GetEmail(ctx context.Context, email string) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	PutUser(ctx context.Context, user User) (string, error)
	RemoveUser(ctx context.Context, id string) error
}

// A PolicyStore knows how to store and retrieve skribe Policies.
type PolicyStore interface {
	GetPolicy(ctx context.Context, id string) (Policy, error)
	PutPolicy(ctx context.Context, policy Policy) (string, error)
	RemovePolicy(ctx context.Context, id string) (string, error)
}

// A GroupStore knows how to store and retrieve skribe Groups.
type GroupStore interface {
	GetGroup(ctx context.Context, id string) (Group, error)
	PutGroup(ctx context.Context, group Group) (string, error)
	RemoveGroup(ctx context.Context, id string) error
}

// An Authenticator knows how to authenticate user credentials.
type Authenticator interface {
	Authenticate(ctx context.Context, email, token string)
}

// A FileStore knows how to store and retrieve skribe document contents.
type FileStore interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	RemoveFile(path string) error
	ListDir(path string) ([]string, error)
}
