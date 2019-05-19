package main

import (
	"log"
	"net"

	"google.golang.org/grpc/reflection"

	"github.com/sakjur/keyfob"
	"github.com/sakjur/keyfob/bolt"
	"github.com/sakjur/keyfob/proto"
	"google.golang.org/grpc"
)

func main() {
	vault, err := bolt.NewVault()
	if err != nil {
		log.Panic(err)
	}

	s := grpc.NewServer()
	proto.RegisterKeyFobServiceServer(s, &keyfob.KeyFobService{
		Vault: vault,
	})
	reflection.Register(s)

	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Panic(err)
	}

	err = s.Serve(listener)
	if err != nil {
		log.Panic(err)
	}
}
