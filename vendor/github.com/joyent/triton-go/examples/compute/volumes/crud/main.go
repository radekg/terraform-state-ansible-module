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
	"time"

	"github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
)

func main() {
	keyID := os.Getenv("TRITON_KEY_ID")
	accountName := os.Getenv("TRITON_ACCOUNT")
	keyMaterial := os.Getenv("TRITON_KEY_MATERIAL")
	userName := os.Getenv("TRITON_USER")

	var signer authentication.Signer
	var err error

	if keyMaterial == "" {
		input := authentication.SSHAgentSignerInput{
			KeyID:       keyID,
			AccountName: accountName,
			Username:    userName,
		}
		signer, err = authentication.NewSSHAgentSigner(input)
		if err != nil {
			log.Fatalf("Error Creating SSH Agent Signer: %v", err)
		}
	} else {
		var keyBytes []byte
		if _, err = os.Stat(keyMaterial); err == nil {
			keyBytes, err = ioutil.ReadFile(keyMaterial)
			if err != nil {
				log.Fatalf("Error reading key material from %s: %v",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				log.Fatalf(
					"Failed to read key material %q: no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				log.Fatalf(
					"Failed to read key %q: password protected keys are\n"+
						"not currently supported. Please decrypt the key prior to use.", keyMaterial)
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
			log.Fatalf("Error Creating SSH Private Key Signer: %v", err)
		}
	}

	config := &triton.ClientConfig{
		TritonURL:   os.Getenv("TRITON_URL"),
		AccountName: accountName,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	c, err := compute.NewClient(config)
	if err != nil {
		log.Fatalf("compute.NewClient: %v", err)
	}

	startTime := time.Now()
	volume, err := c.Volumes().Create(context.Background(), &compute.CreateVolumeInput{
		Type:     "tritonnfs",
		Size:     10240,
		Name:     "my-test-volume-1",
		Networks: []string{"d251d640-a02e-47b5-8ae6-8b45d859528e"},
	})
	if err != nil {
		log.Fatalf("compute.Volumes.Create: %v", err)
	}

	// Wait for provisioning to complete...
	state := make(chan *compute.Volume, 1)
	go func(volumeID string, c *compute.ComputeClient) {
		for {
			time.Sleep(1 * time.Second)
			volume, err := c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
				ID: volumeID,
			})
			if err != nil {
				log.Fatalf("Get(): %v", err)
			}
			if volume.State == "ready" {
				state <- volume
			} else {
				fmt.Print(".")
			}
		}
	}(volume.ID, c)

	select {
	case volume := <-state:
		fmt.Printf("\nDuration: %s\n", time.Since(startTime))
		fmt.Println("Name:", volume.Name)
		fmt.Println("State:", volume.State)
	case <-time.After(5 * time.Minute):
		fmt.Println("Timed out")
	}

	err = c.Volumes().Update(context.Background(), &compute.UpdateVolumeInput{
		Name: "my-test-volume-1-updated",
		ID:   volume.ID,
	})
	if err != nil {
		log.Fatalf("Failed to update name of volume %q: %v", volume.ID, err)
	}

	volume, err = c.Volumes().Get(context.Background(), &compute.GetVolumeInput{
		ID: volume.ID,
	})
	if err != nil {
		log.Fatalf("Failed to retrieve details of volume %q: %v", volume.ID, err)
	}

	fmt.Println("Name:", volume.Name)
	fmt.Println("State:", volume.State)

	fmt.Println("\nCleaning up volume....")
	err = c.Volumes().Delete(context.Background(), &compute.DeleteVolumeInput{
		ID: volume.ID,
	})
	if err != nil {
		log.Fatalf("Failed to delete volume %q: %v", volume.ID, err)
	}
}
