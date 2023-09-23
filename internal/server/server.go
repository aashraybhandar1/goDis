package server

import (
	"context"

	api "github.com/aashraybhandar1/goDis/internal/api/v1"
	"google.golang.org/grpc"
)

// CommitLog an interface that with function Read,Append
type Config struct {
	CommitLog CommitLog
}

// Not sure what this is
var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{
		Config: config,
	}
	return srv, nil
}

// Create a NewGRPCServer given a config. Returns grpc.NewServer()
func NewGRPCServer(config *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}

// Implementing functions of api.UnimplementedLogServer.
// Each method gets an additional request parameter ctx
// Since commitLog is an interface any struct that implements Read and Append are an implementation of this interface. i.e log.go
// Fetch outfit by calling the method implementation Append
func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (
	*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

// Implementing functions of api.UnimplementedLogServer.
// Each method gets an additional request parameter ctx
// Since commitLog is an interface any struct that implements Read and Append are an implementation of this interface. i.e log.go
// Fetch outfit by calling the method implementation Read
func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (
	*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{Record: record}, nil
}

// Implementing functions of api.UnimplementedLogServer.
// Each method gets an additional request parameter ctx
// Produce Request contains of a sibgle bi directional stream to send requests and write response
func (s *grpcServer) ProduceStream(
	stream api.Log_ProduceStreamServer,
) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

// / Implementing functions of api.UnimplementedLogServer.
// Each method gets an additional request parameter ctx
// Consume stream has a stream for response where you write the response even future responses
func (s *grpcServer) ConsumeStream(
	req *api.ConsumeRequest,
	stream api.Log_ConsumeStreamServer,
) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}
