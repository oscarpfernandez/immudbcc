package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	immuapi "github.com/codenotary/immudb/pkg/api"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	"github.com/oscarpfernandez/immudbcc/pkg/server"
)

func main() {
	dbServer, err := server.New(server.Config{AuthEnabled: false, LogFile: "immuserver.log"})

	log.Print("Starting ImmuDB Server...")
	dbServer.Start()
	defer func() {
		if err := dbServer.Stop(); err != nil {
			log.Fatalf("Failed to stop server: %v", err)
		}
		log.Print("Stopped ImmuDB Server")
	}()

	time.Sleep(100 * time.Millisecond)

	options := immuclient.DefaultOptions().WithAuth(false)
	client, err := immuclient.NewImmuClient(options)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	key2, value2 := []byte("client/employee/name/string"), []byte("MasterCard Baller 12/19")
	verifiedIndex, err := client.SafeSet(ctx, key2, value2)
	if err != nil {
		exit(err)
	}
	fmt.Println("   SafeSet - add and verify entry:")
	printItem(key2, value2, verifiedIndex)

	key3, value3 := []byte("client/employee/name/string"), []byte("MasterCard 2232703813463070 12/19")
	verifiedIndex, err = client.SafeSet(ctx, key3, value3)
	if err != nil {
		exit(err)
	}
	fmt.Println("   SafeSet - add and verify entry:")
	printItem(key3, value3, verifiedIndex)

	value3 = []byte("MasterCard 8069498678459876 10/22")
	verifiedIndex, err = client.SafeSet(ctx, key3, value3)
	if err != nil {
		exit(err)
	}
	fmt.Println("   SafeSet - update and verify entry:")
	printItem(key3, value3, verifiedIndex)

	//------> SafeReference
	key3Ref := append([]byte("reference:"), key3...)
	verifiedIndex, err = client.SafeReference(ctx, key3Ref, key3)
	if err != nil {
		exit(err)
	}
	fmt.Println("   SafeReference - add and verify a reference key to an existing entry:")
	printItem(key3Ref, value3, verifiedIndex)

	//------> Scan
	prefix := []byte("client:Ms.")
	structuredItemList, err := client.Scan(ctx, prefix)
	if err != nil {
		exit(err)
	}
	fmt.Printf("   Scan - iterate over keys having the specified prefix (e.g. \"%s\"):\n", prefix)
	for _, item := range structuredItemList.Items {
		printItem(nil, nil, item)
		fmt.Println("	------")
	}

}

func printItem(key []byte, value []byte, message interface{}) {
	var index uint64
	ts := uint64(time.Now().Unix())
	var verified, isVerified bool
	var hash []byte
	switch m := message.(type) {
	case *immuschema.Index:
		index = m.Index
		dig := immuapi.Digest(index, key, value)
		hash = dig[:]
	case *immuclient.VerifiedIndex:
		index = m.Index
		dig := immuapi.Digest(index, key, value)
		hash = dig[:]
		verified = m.Verified
		isVerified = true
	case *immuschema.Item:
		key = m.Key
		value = m.Value
		index = m.Index
		hash = m.Hash()
	case *immuschema.StructuredItem:
		key = m.Key
		value = m.Value.Payload
		ts = m.Value.Timestamp
		index = m.Index
		hash, _ = m.Hash()
	case *immuclient.VerifiedItem:
		key = m.Key
		value = m.Value
		index = m.Index
		ts = m.Time
		verified = m.Verified
		isVerified = true
		me, _ := immuschema.Merge(value, ts)
		dig := immuapi.Digest(index, key, me)
		hash = dig[:]

	}
	if !isVerified {
		fmt.Printf("	index:		%d\n	key:		%s\n	value:		%s\n	hash:		%x\n	time:		%s\n",
			index,
			key,
			value,
			hash,
			time.Unix(int64(ts), 0))
		return
	}
	fmt.Printf("	index:		%d\n	key:		%s\n	value:		%s\n	hash:		%x\n	time:		%s\n	verified:	%t\n",
		index,
		key,
		value,
		hash,
		time.Unix(int64(ts), 0),
		verified)
}

func exit(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
