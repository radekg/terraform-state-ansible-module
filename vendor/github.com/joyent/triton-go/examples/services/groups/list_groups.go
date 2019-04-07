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
	"github.com/joyent/triton-go/services"
)

func main() {
	keyID := triton.GetEnv("KEY_ID")
	accountName := triton.GetEnv("ACCOUNT")
	keyMaterial := triton.GetEnv("KEY_MATERIAL")
	userName := triton.GetEnv("USER")
	tritonURL := triton.GetEnv("URL")

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

	if err = os.Setenv("TRITON_TSG_URL", "http://localhost:3000/"); err != nil {
		log.Fatal("failed to set TRITON_TSG_URL")
	}

	config := &triton.ClientConfig{
		AccountName: accountName,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
		TritonURL:   tritonURL,
	}

	svc, err := services.NewClient(config)
	if err != nil {
		log.Fatalf("failed to create new services client: %v", err)
	}

	fmt.Println("---")

	listInput := &services.ListGroupsInput{}
	groups, err := svc.Groups().List(context.Background(), listInput)
	if err != nil {
		log.Fatalf("failed to list service groups: %v", err)
	}

	for _, grp := range groups {
		fmt.Printf("Group ID: %v\n", grp.ID)
		fmt.Printf("Group Name: %v\n", grp.GroupName)
		fmt.Printf("Group TemplateID: %v\n", grp.TemplateID)
		fmt.Printf("Group Capacity: %v\n", grp.Capacity)
		fmt.Println("")
	}

	fmt.Println("---")

	createTmplInput := &services.CreateTemplateInput{
		TemplateName: "custom-template-1",
		Package:      "test-package",
		ImageID:      "test-image-id",
	}
	tmpl, err := svc.Templates().Create(context.Background(), createTmplInput)
	if err != nil {
		log.Fatalf("failed to create template")
	}

	createInput := &services.CreateGroupInput{
		GroupName:  "custom-group-1",
		TemplateID: tmpl.ID,
		Capacity:   2,
	}
	grp, err := svc.Groups().Create(context.Background(), createInput)
	if err != nil {
		log.Fatalf("failed to create service group: %v", err)
	}

	fmt.Printf("Created Group ID: %s\n", grp.ID)

	fmt.Println("---")

	deleteInput := &services.DeleteGroupInput{
		ID: grp.ID,
	}
	err = svc.Groups().Delete(context.Background(), deleteInput)
	if err != nil {
		log.Fatalf("failed to delete service group: %v", err)
	}

	fmt.Printf("Delete Group: %s\n", grp.GroupName)
}
