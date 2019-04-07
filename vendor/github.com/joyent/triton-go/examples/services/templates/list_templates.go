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
		TritonURL:   tritonURL,
		Username:    userName,
		Signers:     []authentication.Signer{signer},
	}

	svc, err := services.NewClient(config)
	if err != nil {
		log.Fatalf("failed to create new services client: %v", err)
	}

	listInput := &services.ListTemplatesInput{}
	templates, err := svc.Templates().List(context.Background(), listInput)
	if err != nil {
		log.Fatalf("failed to list instance templates: %v", err)
	}
	for _, template := range templates {
		fmt.Printf("Template Name: %s\n", template.TemplateName)
		fmt.Printf("ID: %v\n", template.ID)
		fmt.Printf("Package: %v\n", template.Package)
		fmt.Printf("ImageID: %v\n", template.ImageID)
		fmt.Printf("FirewallEnabled: %v\n", template.FirewallEnabled)
		fmt.Printf("Networks: %v\n", template.Networks)
		fmt.Printf("Userdata: %v\n", template.Userdata)
		fmt.Printf("Metadata: %v\n", template.Metadata)
		fmt.Printf("Tags: %v\n", template.Tags)
		fmt.Println("")
	}

	fmt.Println("---")

	if len(templates) > 0 {
		if tmpl := templates[0]; tmpl != nil {
			getInput := &services.GetTemplateInput{
				ID: tmpl.ID,
			}
			template, err := svc.Templates().Get(context.Background(), getInput)
			if err != nil {
				log.Fatalf("failed to get instance template: %v", err)
			}

			fmt.Printf("Got Template: %s\n", template.TemplateName)
		}
	}

	fmt.Println("---")

	customTemplateName := "custom-template-2"

	createInput := &services.CreateTemplateInput{
		TemplateName:    customTemplateName,
		Package:         "test-package",
		ImageID:         "49b22aec-0c8a-11e6-8807-a3eb4db576ba",
		FirewallEnabled: false,
		Networks:        []string{"f7ed95d3-faaf-43ef-9346-15644403b963"},
		Userdata:        "bash script here",
		Metadata:        map[string]string{"metadata": "test"},
		Tags:            map[string]string{"tag": "test"},
	}
	newTmpl, err := svc.Templates().Create(context.Background(), createInput)
	if err != nil {
		log.Fatalf("failed to create instance template: %v", err)
	}

	fmt.Printf("Created ID: %s\n", newTmpl.ID)
	fmt.Printf("Created TemplateName: %s\n", newTmpl.TemplateName)

	fmt.Println("---")

	deleteInput := &services.DeleteTemplateInput{
		ID: newTmpl.ID,
	}
	err = svc.Templates().Delete(context.Background(), deleteInput)
	if err != nil {
		log.Fatalf("failed to delete instance template: %v", err)
	}

	fmt.Printf("Delete Template: %s\n", customTemplateName)

}
