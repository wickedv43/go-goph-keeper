package server

import (
	"context"

	"github.com/pkg/errors"
	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	user, err := s.service.UserByLogin(ctx, in.Login)
	if err != nil {
		return nil, status.Error(codes.NotFound, "пользователь не найден")
	}

	if user.PasswordHash != in.Password {
		return nil, status.Error(codes.Unauthenticated, "неверный пароль")
	}

	token, err := generateJWT(user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка генерации токена: %v", err)
	}

	s.log.Debugf("userID: ", int64(user.ID))
	return &pb.LoginResponse{
		Token: token,
	}, nil
}

func (s *Server) CreateVault(ctx context.Context, in *pb.CreateVaultRequest) (*emptypb.Empty, error) {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "нет айди юзера")
	}

	v := &storage.VaultRecord{
		UserID:        userID,
		Type:          storage.RecordType(in.Record.Type),
		Title:         in.Record.Title,
		Metadata:      in.Record.Metadata,
		EncryptedData: in.Record.EncryptedData,
	}
	if err = s.service.CreateVault(ctx, v); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось создать запись: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetVault(ctx context.Context, in *pb.GetVaultRequest) (*pb.VaultRecord, error) {
	v, err := s.service.GetVault(ctx, in.VaultId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "запись не найдена: %v", err)
	}
	return mapVaultToProto(&v), nil
}

func (s *Server) UpdateVault(ctx context.Context, in *pb.VaultRecord) (*emptypb.Empty, error) {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "нет айди юзера")
	}

	v := &storage.VaultRecord{
		ID:            in.Id,
		UserID:        userID,
		Type:          storage.RecordType(in.Type),
		Title:         in.Title,
		Metadata:      in.Metadata,
		EncryptedData: in.EncryptedData,
	}
	if err = s.service.UpdateVault(ctx, v); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить запись: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteVault(ctx context.Context, in *pb.DeleteVaultRequest) (*emptypb.Empty, error) {

	if err := s.service.DeleteVault(ctx, in.VaultId); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить запись: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListVaults(ctx context.Context, _ *pb.ListVaultsRequest) (*pb.ListVaultsResponse, error) {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "нет айди юзера")
	}

	records, err := s.service.ListVaults(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить список: %v", err)
	}

	var result []*pb.VaultRecord
	for _, r := range records {
		result = append(result, mapVaultToProto(&r))
	}

	return &pb.ListVaultsResponse{
		Vaults: result,
	}, nil
}
