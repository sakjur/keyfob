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

var defaultServiceKey = "000102030405060708090a0b0c0d0e0f"
var defaultUserID = "7F5CB5F1-32E7-4FD5-87CA-D366617624F6"

var operations = map[string]func(c proto.KeyFobServiceClient, userid string, category string, servicekey string){
	"create": fnCreateKey(),
	"get":    fnListKeys(),
}

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Panic(err)
	}

	userid := flag.String("userid", defaultUserID, "UUID for a user")
	category := flag.String("category", "none", "Category")
	serviceKey := flag.String("servicekey", defaultServiceKey, "Unique key for the calling service")

	flag.Parse()
	op := flag.Arg(0)

	operation, found := operations[op]
	if !found {
		fmt.Printf("Usage: fobc [create|get]\nGot: fobc %s", op)
		return
	}

	fob := proto.NewKeyFobServiceClient(conn)

	operation(fob, *userid, *category, *serviceKey)
}

func fnCreateKey() func(c proto.KeyFobServiceClient, userid string, namespace string, servicekey string) {
	return func(c proto.KeyFobServiceClient, userid string, category string, servicekey string) {
		u, err := uuid.Parse(userid)
		if err != nil {
			log.Panic(err)
		}

		req := &proto.GenerateKeyRequest{
			UserUuid:   u[:],
			Category:   category,
			ServiceKey: []byte(servicekey),
		}

		key, err := c.GenerateKey(context.Background(), req)
		if err != nil {
			log.Panic("Got error on key generation: ", err)
		}

		log.Printf("DERIVED KEY [%s] %x", key.Category, key.Key)
	}
}

func fnListKeys() func(c proto.KeyFobServiceClient, userid string, category string, servicekey string) {
	return func(c proto.KeyFobServiceClient, userid string, category string, servicekey string) {
		u, err := uuid.Parse(userid)
		if err != nil {
			log.Panic(err)
		}

		req := &proto.ListKeysRequest{
			UserUuid:   u[:],
			ServiceKey: serviceKeyBytes(servicekey),
		}

		keys, err := c.ListKeys(context.Background(), req)
		if err != nil {
			log.Panic(err)
		}

		if len(keys.Keys) == 0 {
			fmt.Print("No key found")
			return
		}

		for _, key := range keys.Keys {
			fmt.Printf("DERIVED KEY [%s] %x\n", key.Category, key.Key)
		}
	}
}

func serviceKeyBytes(servicekey string) []byte {
	kb := []byte(servicekey)
	bytes := make([]byte, 2*len(kb))
	_, err := hex.Decode(bytes, kb)
	if err != nil {
		log.Panic(err)
	}
	return bytes
}
