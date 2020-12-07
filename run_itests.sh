#!/usr/bin/env bash

NC='\033[0m'        # Text Reset
YELLOW='\033[0;93m' # Yellow
GREEN='\033[0;92m'  # Green

function echoOK() {
  echo -e "${NC}${GREEN}${1}${NC}"
}

function echoWarning() {
  echo -e "${NC}${YELLOW}${1}${NC}"
}

go build -ldflags "-s -w" -mod vendor -v github.com/oscarpfernandez/immudbcc/cmd/immudb-doc/...

for filename in ./testdata/*.json; do
    # Write the JSON document in the database.
    echoWarning "*** Storing document: ${filename}"

    # Generate random document ID.
    docID=$(uuidgen)

    ./immudb-doc write -doc-id "${docID}" -input-json "${filename}"

    # Read the JSON document from the database.
    echoWarning "*** Retrieving document: ${filename}"
    ./immudb-doc read -doc-id "${docID}" -output-json result.json

    # Compare the retrieved JSON document with original one.
    diff <(jq -S . result.json) <(jq -S . "${filename}") > diff.txt

    if [ -s diff.txt ]
    then
        echo "****************************************************************"
        echo "* Failed to Store and Retrieve Document."
        echo "* Payload did NOT matched (${filename})"
        echo "****************************************************************"
        exit 1
    else
        echoOK "**************************************************************"
        echoOK "* Successfully Store and Retrieve Document."
        echoOK "* Payload matched for file (${filename})"
        echoOK "**************************************************************"
    fi
done