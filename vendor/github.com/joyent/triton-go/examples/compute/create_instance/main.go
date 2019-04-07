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

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/compute"
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

const (
	PackageName  = "g4-highcpu-512M"
	ImageName    = "ubuntu-16.04"
	ImageVersion = "20170403"
	NetworkName  = "Joyent-SDC-Public"
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
				log.Fatalf("Error reading key material from %s: %s",
					keyMaterial, err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				log.Fatalf(
					"Failed to read key material '%s': no key found", keyMaterial)
			}

			if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
				log.Fatalf(
					"Failed to read key '%s': password protected keys are\n"+
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
		log.Fatalf("Compute NewClient(): %v", err)
	}
	n, err := network.NewClient(config)
	if err != nil {
		log.Fatalf("Network NewClient(): %v", err)
	}

	images, err := c.Images().List(context.Background(), &compute.ListImagesInput{
		Name:    ImageName,
		Version: ImageVersion,
	})
	if err != nil {
		log.Fatalf("compute.Images.List: %v", err)
	}
	var img compute.Image
	if len(images) > 0 {
		img = *images[0]
	} else {
		log.Fatalf("Unable to find an Image")
	}

	var net *network.Network
	nets, err := n.List(context.Background(), &network.ListInput{})
	if err != nil {
		log.Fatalf("Network List(): %v", err)
	}
	for _, found := range nets {
		if found.Name == NetworkName {
			net = found
		}
	}

	// Create a new instance using our input attributes...
	// https://github.com/joyent/triton-go/blob/master/compute/instances.go#L206
	createInput := &compute.CreateInstanceInput{
		Name:     testutils.RandString(10),
		Package:  PackageName,
		Image:    img.ID,
		Networks: []string{net.Id},
		Metadata: map[string]string{
			"user-script": "<your script here>",
		},
		Tags: map[string]string{
			"tag1": "value1",
		},
		CNS: compute.InstanceCNS{
			Services: []string{"frontend", "web"},
		},
	}
	startTime := time.Now()
	created, err := c.Instances().Create(context.Background(), createInput)
	if err != nil {
		log.Fatalf("Create(): %v", err)
	}

	// Wait for provisioning to complete...
	state := make(chan *compute.Instance, 1)
	go func(createdID string, c *compute.ComputeClient) {
		for {
			time.Sleep(1 * time.Second)
			instance, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: createdID,
			})
			if err != nil {
				log.Fatalf("Get(): %v", err)
			}
			if instance.State == "running" {
				state <- instance
			} else {
				fmt.Print(".")
			}
		}
	}(created.ID, c)

	select {
	case instance := <-state:
		fmt.Printf("\nDuration: %s\n", time.Since(startTime))
		fmt.Println("Name:", instance.Name)
		fmt.Println("State:", instance.State)
	case <-time.After(5 * time.Minute):
		fmt.Println("Timed out")
	}

	fmt.Println("Cleaning up machine....")
	err = c.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
		ID: created.ID,
	})
	if err != nil {
		log.Fatalf("Delete(): %v", err)
	}
}
