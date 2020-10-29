package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	immuapi "github.com/codenotary/immudb/pkg/api"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	immulogger "github.com/codenotary/immudb/pkg/logger"
	immuserver "github.com/codenotary/immudb/pkg/server"
)

func main() {
	fmt.Println("1. Start immudb server ...")
	const logfile = "immuserver.log"
	flogger, file, err :=
		immulogger.NewFileLogger("immuserver ", logfile)
	if err != nil {
		exit(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			exit(err)
		}
	}()
	serverOptions := immuserver.DefaultOptions().WithLogfile(logfile).WithAuth(false)
	server := immuserver.DefaultServer().WithOptions(serverOptions).WithLogger(flogger)
	go func() {
		if err := server.Start(); err != nil {
			exit(err)
		}
	}()
	defer func() {
		err := server.Stop()
		// NOTE: this cleanup must NOT be done in a real-world scenario!
		cleanup(serverOptions.Dir, serverOptions.Logfile)
		if err != nil {
			exit(err)
		}
	}()
	// wait for server to start
	time.Sleep(100 * time.Millisecond)

	options := immuclient.DefaultOptions()
	options.Auth = false
	client, err := immuclient.NewImmuClient(options)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	key3, value3 := []byte("client:Ms. Maci Schuppe"), []byte("MasterCard 2232703813463070 12/19")
	verifiedIndex, err := client.SafeSet(ctx, key3, value3)
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

func cleanup(dbDir string, logfile string) {
	// remove db
	os.RemoveAll(dbDir)
	// remove log file
	os.Remove(logfile)
	// remove root
	files, err := filepath.Glob("./\\.root*")
	if err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}
}
