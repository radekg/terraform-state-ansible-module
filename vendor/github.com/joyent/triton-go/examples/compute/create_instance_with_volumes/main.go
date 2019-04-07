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
	"github.com/joyent/triton-go/network"
	"github.com/joyent/triton-go/testutils"
)

const (
	PackageName       = "g4-highcpu-512M"
	ImageName         = "ubuntu-16.04"
	ImageVersion      = "20170403"
	PublicNetworkName = "Joyent-SDC-Public"
	PrivateWorkName   = "My-Fabric-Network"
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

	n, err := network.NewClient(config)
	if err != nil {
		log.Fatalf("Network NewClient(): %v", err)
	}

	var publicNet *network.Network
	var privateNet *network.Network
	nets, err := n.List(context.Background(), &network.ListInput{})
	if err != nil {
		log.Fatalf("Network List(): %v", err)
	}
	for _, found := range nets {
		if found.Name == PublicNetworkName {
			publicNet = found
		} else if found.Name == PrivateWorkName {
			privateNet = found
		}
	}

	volume, err := c.Volumes().Create(context.Background(), &compute.CreateVolumeInput{
		Type:     "tritonnfs",
		Size:     10240,
		Name:     testutils.RandString(10),
		Networks: []string{privateNet.Id},
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
				log.Fatalf("compute.Volumes.Get(): %v", err)
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
		fmt.Println("\n#### Volume")
		fmt.Println("Name:", volume.Name)
		fmt.Println("State:", volume.State)
	case <-time.After(5 * time.Minute):
		fmt.Println("\nTimed out")
	}

	// Now we create the machine with the new volume we created above
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

	instanceVolume := compute.InstanceVolume{
		Name:       volume.Name,
		Type:       volume.Type,
		Mountpoint: "/foo",
		Mode:       "rw",
	}

	// Create a new instance using our input attributes...
	// https://github.com/joyent/triton-go/blob/master/compute/instances.go#L206
	createInput := &compute.CreateInstanceInput{
		Name:     testutils.RandString(10),
		Package:  PackageName,
		Image:    img.ID,
		Networks: []string{publicNet.Id, privateNet.Id},
		Volumes: []compute.InstanceVolume{
			instanceVolume,
		},
		Tags: map[string]string{
			"tag1": "value1",
		},
	}

	created, err := c.Instances().Create(context.Background(), createInput)
	if err != nil {
		log.Fatalf("compute.Instances.Create: %v", err)
	}

	// Wait for provisioning to complete...
	newState := make(chan *compute.Instance, 1)
	go func(createdID string, c *compute.ComputeClient) {
		for {
			time.Sleep(1 * time.Second)
			instance, err := c.Instances().Get(context.Background(), &compute.GetInstanceInput{
				ID: createdID,
			})
			if err != nil {
				log.Fatalf("compute.Instances.Get: %v", err)
			}
			if instance.State == "running" {
				newState <- instance
			} else {
				fmt.Print(".")
			}
		}
	}(created.ID, c)

	select {
	case machine := <-newState:
		fmt.Println("\n\n#### Instance")
		fmt.Println("Name:", machine.Name)
		fmt.Println("State:", machine.State)
	case <-time.After(5 * time.Minute):
		fmt.Println("Timed out")
	}

	fmt.Println("\nCleaning up machine....")
	err = c.Instances().Delete(context.Background(), &compute.DeleteInstanceInput{
		ID: created.ID,
	})
	if err != nil {
		log.Fatalf("compute.Instances.Delete(): %v", err)
	}

	time.Sleep(30 * time.Second)

	fmt.Println("\nCleaning up volume....")
	err = c.Volumes().Delete(context.Background(), &compute.DeleteVolumeInput{
		ID: volume.ID,
	})
	if err != nil {
		log.Fatalf("compute.Volumes.Delete(): %v", err)
	}
}
