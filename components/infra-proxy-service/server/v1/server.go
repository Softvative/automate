package v1

import (
	"context"

	uuid "github.com/chef/automate/lib/uuid4"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/chef/automate/api/interservice/infra_proxy/request"
	"github.com/chef/automate/api/interservice/infra_proxy/response"

	"github.com/chef/automate/components/infra-proxy-service/service"
	"github.com/chef/automate/components/infra-proxy-service/storage"
)

// CreateServer creates a new server
func (s *Server) CreateServer(ctx context.Context, req *request.CreateServer) (*response.CreateServer, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if req.Name == "" {
		s.service.Logger.Debug("incomplete create server request: missing server name")
		return nil, status.Error(codes.InvalidArgument, "must supply sever name")
	}

	if req.Fqdn == "" {
		s.service.Logger.Debug("incomplete create server request: missing server fqdn")
		return nil, status.Error(codes.InvalidArgument, "must supply server fqdn")
	}

	var server storage.Server
	var err error
	if server, err = s.service.Storage.StoreServer(ctx, req.Name, req.Fqdn, req.IpAddress); err != nil {
		if err == storage.ErrConflict {
			return nil, status.Errorf(codes.AlreadyExists, "server with name %q already exists", req.Name)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &response.CreateServer{
		Server: fromStorageServer(server),
	}, nil
}

// GetServers returns a list of servers from the db
func (s *Server) GetServers(ctx context.Context, req *request.GetServers) (*response.GetServers, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	serversList, err := s.service.Storage.GetServers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &response.GetServers{
		Servers: fromStorageToListServers(serversList),
	}, nil
}

// GetServer takes an ID and returns a server object
func (s *Server) GetServer(ctx context.Context, req *request.GetServer) (*response.GetServer, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	UUID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid server id")
	}

	server, err := s.service.Storage.GetServer(ctx, UUID)
	if err != nil {
		return nil, service.ParseStorageError(err, req.Id, "server")
	}

	return &response.GetServer{
		Server: fromStorageServer(server),
	}, nil
}

// DeleteServer deletes a server from the db
func (s *Server) DeleteServer(ctx context.Context, req *request.DeleteServer) (*response.DeleteServer, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	UUID, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid server id")
	}

	server, err := s.service.Storage.DeleteServer(ctx, UUID)
	if err != nil {
		return nil, service.ParseStorageError(err, req.Id, "server")
	}

	if err != nil {
		s.service.Logger.Warnf("failed to purge subjects on server delete: %s", err.Error())
		return nil, status.Errorf(codes.Internal, "failed to purge server %q from policies: %s", server.ID, err.Error())
	}

	return &response.DeleteServer{
		Server: fromStorageServer(server),
	}, nil
}

// UpdateServer updates a server in the db via post
func (s *Server) UpdateServer(ctx context.Context, req *request.UpdateServer) (*response.UpdateServer, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if req.Id == "" {
		s.service.Logger.Debug("incomplete update server request: missing server id")
		return nil, status.Error(codes.InvalidArgument, "must supply server id")
	}
	if req.Name == "" {
		s.service.Logger.Debug("incomplete update server request: missing server name")
		return nil, status.Error(codes.InvalidArgument, "must supply server name")
	}
	if req.Fqdn == "" {
		s.service.Logger.Debug("incomplete update server request: missing server fqdn")
		return nil, status.Error(codes.InvalidArgument, "must supply server fqdn")
	}
	if req.IpAddress == "" {
		s.service.Logger.Debug("incomplete update server request: missing server ip_address")
		return nil, status.Error(codes.InvalidArgument, "must supply server ip_address")
	}

	id, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid server id")
	}

	serverStruct := storage.Server{
		ID:        id,
		Name:      req.Name,
		Fqdn:      req.Fqdn,
		IpAddress: req.IpAddress,
	}

	server, err := s.service.Storage.EditServer(ctx, serverStruct)
	if err != nil {
		return nil, service.ParseStorageError(err, id, "server")
	}

	return &response.UpdateServer{
		Server: fromStorageServer(server),
	}, nil
}

// Create a response.Server from a storage.Server
func fromStorageServer(s storage.Server) *response.Server {
	return &response.Server{
		Id:        s.ID.String(),
		Name:      s.Name,
		Fqdn:      s.Fqdn,
		IpAddress: s.IpAddress,
	}
}

// Create a response.ServersList from an array of storage.Server
func fromStorageToListServers(sl []storage.Server) []*response.Server {
	tl := make([]*response.Server, len(sl))

	for i, server := range sl {
		tl[i] = fromStorageServer(server)
	}

	return tl
}
