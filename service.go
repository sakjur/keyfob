package keyfob

import (
	"context"

	"github.com/google/uuid"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sakjur/keyfob/proto"
)

type KeyFobService struct {
	Vault KeyVault
}

func (s *KeyFobService) GetKey(ctx context.Context, req *proto.GetKeyRequest) (*proto.EncryptionKey, error) {
	userid, err := uuid.FromBytes(req.RowKey.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKey{UserID: userid, Namespace: req.RowKey.Namespace, ServiceKey: req.ServiceKey}
	key, err := userKey.DeriveKey(s.Vault)
	if err != nil {
		return nil, err
	}

	return &proto.EncryptionKey{
		Key: key,
	}, nil
}

func (s *KeyFobService) GetOrCreateKey(ctx context.Context, req *proto.GetKeyRequest) (*proto.EncryptionKey, error) {
	userid, err := uuid.FromBytes(req.RowKey.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKey{UserID: userid, Namespace: req.RowKey.Namespace, ServiceKey: req.ServiceKey}
	key, err := userKey.DeriveKey(s.Vault)
	if err != nil {
		_ = userKey.CreateKey(s.Vault)
		key, err = userKey.DeriveKey(s.Vault)
		if err != nil {
			return nil, err
		}
	}

	return &proto.EncryptionKey{
		Key: key,
	}, nil
}

func (s *KeyFobService) DeleteKey(ctx context.Context, req *proto.DeleteKeyRequest) (*empty.Empty, error) {
	userid, err := uuid.FromBytes(req.RowKey.UserUuid)
	if err != nil {
		return nil, err
	}

	userKey := UserKey{UserID: userid, Namespace: req.RowKey.Namespace}
	return &empty.Empty{}, userKey.DeleteKey(s.Vault)
}
