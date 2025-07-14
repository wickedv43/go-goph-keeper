package server

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	u := &storage.User{
		Login:        in.Login,
		PasswordHash: in.Password,
	}

	user, err := s.service.NewUser(ctx, u)
	if err != nil {
		if errors.Is(storage.ErrLoginUsed, err) {
			return nil, err
		}
		return nil, err
	}

	s.log.Debugf("new user: " + user.Login)

	return &pb.RegisterResponse{UserId: user.ID}, nil
}

func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.service.FindUserByLogin(ctx, in.Login)
	if err != nil {
		return nil, status.Error(codes.NotFound, "пользователь не найден")
	}

	if user.PasswordHash != in.Password {
		return nil, status.Error(codes.Unauthenticated, "неверный пароль")
	}

	return &pb.LoginResponse{
		UserId: user.ID,
		Login:  user.Login,
	}, nil
}
