package storage

import (
	"context"
	"errors"
	"time"

	uuid "github.com/chef/automate/lib/uuid4"
)

// Storage is the interface provided by our various storage backends.
type Storage interface {
	GetServer(context.Context, uuid.UUID) (Server, error)
	GetServers(context.Context) ([]Server, error)
	StoreServer(ctx context.Context, name string, fqdn string, ipAddress string) (Server, error)
	DeleteServer(context.Context, uuid.UUID) (Server, error)
	EditServer(context.Context, Server) (Server, error)

	GetOrg(context.Context, uuid.UUID) (Org, error)
	GetOrgs(context.Context) ([]Org, error)
	StoreOrg(ctx context.Context, name string, adminUser string, adminKey string, serverId string) (Org, error)
	DeleteOrg(context.Context, uuid.UUID) (Org, error)
	EditOrg(context.Context, Org) (Org, error)
}

// Resetter is, if exposed, used for tests to reset the storage backend to a
// pristine state.
type Resetter interface {
	Reset(context.Context) error
}

// Server is the struct ingested and returned by our backend implementations.
type Server struct {
	ID        uuid.UUID
	Name      string
	Fqdn      string
	IpAddress string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Org is the struct ingested and returned by our backend implementations.
type Org struct {
	ID        uuid.UUID
	Name      string
	AdminUser string
	AdminKey  string
	ServerId  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Errors returned from the backend
var (
	// ErrNotFound is returned when a requested server wasn't found
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned when a server there is a clash of server IDs
	ErrConflict = errors.New("conflict")
)
