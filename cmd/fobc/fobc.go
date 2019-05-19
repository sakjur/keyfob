package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/sakjur/keyfob/proto"
	"google.golang.org/grpc"
)

var serviceKey = "000102030405060708090a0b0c0d0e0f"
var defaultUserID = "7F5CB5F1-32E7-4FD5-87CA-D366617624F6"

var operations = map[string]func(c proto.KeyFobServiceClient, userid string, namespace string, servicekey string){
	"create": fnGetKey(true),
	"get":    fnGetKey(false),
}

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Panic(err)
	}

	userid := flag.String("userid", defaultUserID, "UUID for a user")
	namespace := flag.String("namespace", "none", "Namespace")
	serviceKey := flag.String("servicekey", serviceKey, "Unique key for the calling service")

	flag.Parse()
	op := flag.Arg(0)

	operation, found := operations[op]
	if !found {
		fmt.Printf("Usage: fobc [create|get]\nGot: fobc %s", op)
		return
	}

	fob := proto.NewKeyFobServiceClient(conn)

	operation(fob, *userid, *namespace, *serviceKey)
}

func fnGetKey(createIfNotExists bool) func(c proto.KeyFobServiceClient, userid string, namespace string, servicekey string) {
	return func(c proto.KeyFobServiceClient, userid string, namespace string, servicekey string) {
		req := &proto.GetKeyRequest{RowKey: &proto.Row{
			Namespace: namespace,
		}}

		u, err := uuid.Parse(userid)
		if err != nil {
			log.Panic(err)
		}

		req.RowKey.UserUuid = u[:]

		kb := []byte(servicekey)
		req.ServiceKey = make([]byte, 2*len(kb))
		_, err = hex.Decode(req.ServiceKey, kb)
		if err != nil {
			log.Panic(err)
		}

		var key *proto.EncryptionKey
		if createIfNotExists {
			key, err = c.GetOrCreateKey(context.Background(), req)
		} else {
			key, err = c.GetKey(context.Background(), req)
		}
		if err != nil {
			log.Panic(err)
		}

		if len(key.Key) == 0 {
			fmt.Print("No key found")
			return
		}

		fmt.Printf("Got key: %x", key.Key)
	}
}
