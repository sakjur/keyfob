package keyfob

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/sakjur/keyfob/proto"
	"golang.org/x/net/context"
)

// KeyFobService implements the KeyFob gRPC service. All keys which are
// returned are derived from both the key stored in the vault and the provided
// service key.
type KeyFobService struct {
	Vault KeyVault
}

// GenerateKey fetches a key from the vault matching the key in the request,
// creating it if it doesn't exist.
func (s KeyFobService) GenerateKey(ctx context.Context, req *proto.GenerateKeyRequest) (*proto.EncryptionKey, error) {
	userid, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKeyPointer{UserID: userid, Category: req.Category, ServiceKey: req.ServiceKey}
	key, err := userKey.DeriveKey(s.Vault)
	if err != nil {
		_ = userKey.CreateKey(s.Vault)
		key, err = userKey.DeriveKey(s.Vault)
		if err != nil {
			return nil, err
		}
	}

	return &proto.EncryptionKey{
		Category: req.Category,
		Key:      key,
	}, nil

}

// DeleteKey permanently deletes a key from the vault.
func (s KeyFobService) DeleteKey(ctx context.Context, req *proto.DeleteKeyRequest) (*empty.Empty, error) {
	userid, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKeyPointer{UserID: userid, Category: req.Category}
	return &empty.Empty{}, userKey.DeleteKey(s.Vault)
}

// ListKeys returns all the keys which exists for a user in a vault.
func (s KeyFobService) ListKeys(ctx context.Context, req *proto.ListKeysRequest) (*proto.ListKeysResponse, error) {
	userid, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKeyPointer{UserID: userid, ServiceKey: req.ServiceKey}
	stored, err := userKey.ListUserKeys(s.Vault)

	keys := make([]*proto.EncryptionKey, len(stored))
	for i, key := range stored {
		keys[i] = &proto.EncryptionKey{
			Key:      key.Key,
			Category: key.Category,
		}
	}

	return &proto.ListKeysResponse{
		Keys: keys,
	}, nil
}
