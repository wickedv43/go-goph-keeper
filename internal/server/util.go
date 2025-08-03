package server

import (
	"time"

	pb "github.com/wickedv43/go-goph-keeper/internal/api"
	"github.com/wickedv43/go-goph-keeper/internal/storage"
)

// mapVaultToProto converts a VaultRecord from the storage layer to its protobuf representation.
func mapVaultToProto(v *storage.VaultRecord) *pb.VaultRecord {
	return &pb.VaultRecord{
		Id:            v.ID,
		UserId:        v.UserID,
		Type:          string(v.Type),
		Title:         v.Title,
		Metadata:      v.Metadata,
		EncryptedData: v.EncryptedData,
		CreatedAt:     v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     v.UpdatedAt.Format(time.RFC3339),
	}
}
