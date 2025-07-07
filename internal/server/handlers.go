package server

import (
	"context"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
)

func (s *Server) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	u := &storage.User{
		Login:        in.Email,
		PasswordHash: in.Password,
	}

	user, err := s.service.NewUser(ctx, u)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{UserId: user.ID}, nil
}

func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{}, nil
}
