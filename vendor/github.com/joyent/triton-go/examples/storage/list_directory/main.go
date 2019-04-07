//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package main

import (
	"context"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/storage"
)

const path = "/stor"

func main() {
	var (
		signer authentication.Signer
		err    error

		keyID       = os.Getenv("MANTA_KEY_ID")
		accountName = os.Getenv("MANTA_USER")
		keyMaterial = os.Getenv("MANTA_KEY_MATERIAL")
		userName    = os.Getenv("TRITON_USER")
	)

	if keyMaterial == "" {
		input := authentication.SSHAgentSignerInput{
			KeyID:       keyID,
			AccountName: accountName,
			Username:    userName,
		}
		signer, err = authentication.NewSSHAgentSigner(input)
		if err != nil {
			log.Fatalf("error creating SSH agent signer: %v", err)
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(keyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(keyMaterial)
			if err != nil {
				log.Fatalf("error reading key material from %q: %v",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				log.Fatalf(
					"failed to read key material %q: no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				log.Fatalf("failed to read key %q: password protected keys are\n"+
					"not currently supported, decrypt key prior to use",
					keyMaterial)
			}

		} else {
			keyBytes = []byte(keyMaterial)
		}

		input := authentication.PrivateKeySignerInput{
			KeyID:              keyID,
			PrivateKeyMaterial: keyBytes,
			AccountName:        accountName,
			Username:           userName,
		}
		signer, err = authentication.NewPrivateKeySigner(input)
		if err != nil {
			log.Fatalf("error creating SSH private key signer: %v", err)
		}
	}

	config := &triton.ClientConfig{
		MantaURL:    os.Getenv("MANTA_URL"),
		AccountName: accountName,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	client, err := storage.NewClient(config)
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}

	ctx := context.Background()
	output, err := client.Dir().List(ctx, &storage.ListDirectoryInput{
		DirectoryName: path,
	})
	if err != nil {
		fmt.Printf("could not find %q\n", path)
		return
	}

	for _, item := range output.Entries {
		fmt.Print("******* ITEM *******\n")
		fmt.Printf("Name: %s\n", item.Name)
		fmt.Printf("Type: %s\n", item.Type)
		fmt.Print("\n")
	}
}
