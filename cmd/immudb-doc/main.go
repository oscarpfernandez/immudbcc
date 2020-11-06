package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/oscarpfernandez/immudbcc/pkg/api"
	"github.com/oscarpfernandez/immudbcc/pkg/doc"
	"github.com/oscarpfernandez/immudbcc/pkg/server"

	immuapi "github.com/codenotary/immudb/pkg/api"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
)

func main() {
	fsWrite := flag.NewFlagSet("write", flag.ContinueOnError)
	jsonPath := fsWrite.String("input-json", "", "JSON path of the file to store")

	if len(os.Args) <= 1 {
		fmt.Printf(os.Args[0] + " <read | write>  [flags]\n")
		fmt.Println("* Flags <write>")
		fsWrite.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "write":
		fsWrite.Parse(os.Args[2:])
	case "read":
		// TODO: implements this.
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if fsWrite.Parsed() {
		if *jsonPath == "" {
			fsWrite.PrintDefaults()
			os.Exit(1)
		} else {
			if _, err := os.Stat(*jsonPath); os.IsExist(err) {
				log.Fatalf("File does not exist: %s", err)
			}
		}
	}

	jsonReader, err := openFile(*jsonPath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer jsonReader.Close()

	dbServer, err := server.New(server.Config{AuthEnabled: false, LogFile: "immuserver.log"})
	if err != nil {
		log.Fatalf("Failed to init server: %v", err)
	}

	log.Print("Starting ImmuDB Server...")
	dbServer.Start()
	defer func() {
		if err := dbServer.Stop(); err != nil {
			log.Fatalf("Failed to stop server: %v", err)
		}
		log.Print("Stopped ImmuDB Server")
	}()

	token, err := doc.GenerateEncryptionToken()
	if err != nil {
		log.Fatalf("Failed to create encryption token: %v", err)
	}
	log.Printf("Encryption token: %s", token)

	conf := api.DefaultConfig().WithEncryptionToken(token)
	apiManager, err := api.New(conf)
	if err != nil {
		log.Fatalf("Failed to start API manager: %v", err)
	}

	now := time.Now()
	result, err := apiManager.StoreDocument(context.Background(), "docID", jsonReader)
	if err != nil {
		log.Fatalf("Failed to store document: %v", err)
	}
	execTime := time.Now().Sub(now).String()
	log.Printf("Write document execution time: %s", execTime)

	log.Printf("Result hash: Index(%d), Hash(%s), EncHash(%s)", result.Index, hex.EncodeToString(result.Hash), hex.EncodeToString(result.HashEnc))

}

func openFile(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
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
