package keyfob

import (
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/sakjur/keyfob/proto"
	"golang.org/x/net/context"
)

type KeyFobService struct {
	Vault KeyVault
}

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

func (s KeyFobService) DeleteKey(ctx context.Context, req *proto.DeleteKeyRequest) (*empty.Empty, error) {
	userid, err := uuid.FromBytes(req.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKeyPointer{UserID: userid, Category: req.Category}
	return &empty.Empty{}, userKey.DeleteKey(s.Vault)
}

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
