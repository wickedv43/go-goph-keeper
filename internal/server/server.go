package server

import (
	"net"

	"github.com/samber/do/v2"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"github.com/wickedv43/go-goph-keeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedGophKeeperServer

	// GRPC is the underlying gRPC server instance.
	GRPC *grpc.Server

	// service
	service service.GophKeeper

	cfg *config.Config
	log *zap.SugaredLogger
}

func NewServer(i do.Injector) (*Server, error) {
	s := &Server{
		service: do.MustInvoke[service.GophKeeper](i),
		cfg:     do.MustInvoke[*config.Config](i),
		log:     do.MustInvoke[*logger.Logger](i).Named("server"),
	}

	grpcServer := grpc.NewServer(
	//grpc.UnaryInterceptor(
	//	ChainUnaryInterceptors(
	//		s.LogUnaryInterceptor(),
	//		s.AuthInterceptor(),
	//	),
	//),
	)

	s.GRPC = grpcServer

	pb.RegisterGophKeeperServer(s.GRPC, s)

	return s, nil
}

func (s *Server) Start() {
	s.log.Debugf("grpc server port: " + s.cfg.Server.Port)
	listen, err := net.Listen("tcp", ":"+s.cfg.Server.Port)
	if err != nil {
		s.log.Fatal("failed to listen")
	}

	if err = s.GRPC.Serve(listen); err != nil {
		s.log.Fatal("failed to serve")
	}
}

func (s *Server) Shutdown() {
	s.log.Debug("grpc shutdown complete ")
	s.GRPC.GracefulStop()
}
