package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

var terraformConfigUnderTest string
var terraformLocalStateUnderTest string

func TestMain(m *testing.M) {

	tempStateFile, _ := ioutil.TempFile("", "tempStateFile*.tfstate")
	ioutil.WriteFile(tempStateFile.Name(), []byte(testState), 0644)

	tempTerraformConfigFile, _ := ioutil.TempFile("", "tempTerraformConfigFile*.tf")
	ioutil.WriteFile(tempTerraformConfigFile.Name(), []byte(fmt.Sprintf(testTerraformConfig, tempStateFile.Name())), 0644)

	tempArgsFile, _ := ioutil.TempFile("", "tempArgsFile")
	ioutil.WriteFile(tempArgsFile.Name(), []byte(fmt.Sprintf(testArgs, tempTerraformConfigFile.Name())), 0644)

	terraformConfigUnderTest = tempArgsFile.Name()
	terraformLocalStateUnderTest = tempStateFile.Name()

	defer os.Remove(tempStateFile.Name())
	defer os.Remove(tempTerraformConfigFile.Name())
	defer os.Remove(tempArgsFile.Name())

	result := m.Run()

	os.Exit(result)

}

/*func TestExecution(t *testing.T) {
	executeProgram([]string{"program-name", terraformConfigUnderTest})
}*/

func TestValidRetrieve(t *testing.T) {
	var err error
	err = validateRetrieve("o/output_name")
	if err != nil {
		t.Fatal("Expected no error o/output_name")
	}
	err = validateRetrieve("r/resource/attribute")
	if err != nil {
		t.Fatal("Expected no error for r/resource/attribute")
	}
}

func TestFailsWithInvalidRetrieve(t *testing.T) {
	var err error
	err = validateRetrieve("r/resource.only")
	if err == nil {
		t.Fatal("Expected the validation to fail")
	}
	err = validateRetrieve("r/resource/attribute/wrong")
	if err == nil {
		t.Fatal("Expected the validation to fail")
	}
	err = validateRetrieve("u/unknown/resource")
	if err == nil {
		t.Fatal("Expected the validation to fail")
	}
}

func TestReduceToMaps(t *testing.T) {
	items := map[string]interface{}{
		"aws_s3_bucket.first_bucket.property":        "some value",
		"aws_s3_bucket.first_bucket.other_property":  "some value",
		"aws_s3_bucket.second_bucket.property":       "some value",
		"aws_s3_bucket.second_bucket.other_property": "some value",
		"output_name": "output value",
	}
	into := make(map[string]interface{})
	for k, v := range items {
		into = reduceToMap(strings.Split(k, "."), v, into)
	}

	if m1, ok := into["aws_s3_bucket"]; !ok {
		t.Fatalf("Expected aws_s3_bucket to be in the result")
	} else {

		if m2, ok := m1.(map[string]interface{})["first_bucket"]; !ok {
			t.Fatalf("Expected first_bucket to be in aws_s3_bucket")
		} else {
			if _, ok := m2.(map[string]interface{})["property"]; !ok {
				t.Fatalf("Expected property to be in aws_s3_bucket.first_bucket")
			}
			if _, ok := m2.(map[string]interface{})["other_property"]; !ok {
				t.Fatalf("Expected other_property to be in aws_s3_bucket.first_bucket")
			}
		}

		if m2, ok := m1.(map[string]interface{})["second_bucket"]; !ok {
			t.Fatalf("Expected second_bucket to be in aws_s3_bucket")
		} else {
			if _, ok := m2.(map[string]interface{})["property"]; !ok {
				t.Fatalf("Expected property to be in aws_s3_bucket.second_bucket")
			}
			if _, ok := m2.(map[string]interface{})["other_property"]; !ok {
				t.Fatalf("Expected other_property to be in aws_s3_bucket.second_bucket")
			}
		}

	}

	if _, ok := into["output_name"]; !ok {
		t.Fatalf("Expected output_name to be in the result")
	}
}

func TestPassConfigureKnownBackend(t *testing.T) {
	backend, err := getBackend("local")
	if err != nil {
		t.Fatal("Expected no error")
	}
	if _, err := configureBackend(backend, map[string]interface{}{
		"path": terraformLocalStateUnderTest,
	}); err != nil {
		t.Fatal("Expected no error when configuring a known backend")
	}
}

func TestFailUnknownBackend(t *testing.T) {
	_, err := getBackend("unknown_backend")
	if err == nil {
		t.Fatal("Expected an error")
	}
}

func TestProcessState(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{defaultModulePath},
				Outputs: map[string]*terraform.OutputState{
					"foo": &terraform.OutputState{
						Type:      "string",
						Sensitive: false,
						Value:     "bar",
					},
				},
				Resources: map[string]*terraform.ResourceState{
					"aws_s3_bucket.bucket": &terraform.ResourceState{
						Primary: &terraform.InstanceState{
							Attributes: map[string]string{
								"property1": "value",
							},
						},
					},
				},
			},
		},
	}
	margs := &ModuleArgs{
		TerraformConfigPath: "/not/important/here.tf",
		State:               defaultState,
		Retrieves: []Retrieve{
			Retrieve{
				Retrieve:   "o/foo",
				ModulePath: defaultModulePath,
			},
			Retrieve{
				Retrieve:   "r/aws_s3_bucket.bucket/property1",
				ModulePath: defaultModulePath,
			},
		},
	}
	state.Init()
	data, err := processState(state, margs)
	if err != nil {
		t.Fatalf("Expected no error")
	}
	if _, ok := data["foo"]; !ok {
		t.Fatalf("Expected foo to be in the response data")
	}
	if m1, ok := data["aws_s3_bucket"]; !ok {
		t.Fatalf("Expected aws_s3_bucket to be in the response data")
	} else {
		if m2, ok := m1.(map[string]interface{})["bucket"]; !ok {
			t.Fatalf("Expected bucket to be in the response data aws_s3_bucket")
		} else {
			if _, ok := m2.(map[string]interface{})["property1"]; !ok {
				t.Fatalf("Expected property1 to be in the response data aws_s3_bucket.bucket")
			}
		}
	}
}

func TestFailRequireAllMissingRetrieve(t *testing.T) {
	state := &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{defaultModulePath},
				Outputs: map[string]*terraform.OutputState{
					"foo": &terraform.OutputState{
						Type:      "string",
						Sensitive: false,
						Value:     "bar",
					},
				},
			},
		},
	}
	margs := &ModuleArgs{
		TerraformConfigPath: "/not/important/here.tf",
		State:               defaultState,
		Retrieves: []Retrieve{
			Retrieve{
				Retrieve:   "o/not_foo",
				ModulePath: defaultModulePath,
			},
		},
		RequireAll: true,
	}
	state.Init()
	_, err := processState(state, margs)
	if err == nil {
		t.Fatalf("Expected an error")
	}
}

var testArgs = `{
	"terraform_config_path": "%s",
	"retrieves": [
		{ "retrieve": "o/bucket_backups" },
		{ "retrieve": "r/aws_s3_bucket.backups/bucket_domain_name" },
		{ "retrieve": "o/bucket_internal_backups" },
		{ "retrieve": "r/aws_s3_bucket.internal_backups/bucket_domain_name" },
		{ "retrieve": "r/aws_s3_bucket.internal_backups/request_payer" }
	]
}`

var testTerraformConfig = `terraform {
	backend "local" {
		path = "%s"
	}
}`

var testState = `{
    "version": 3,
    "terraform_version": "0.11.10",
    "serial": 2,
    "lineage": "c79238b2-a8d1-4c84-91be-dcefa88a8884",
    "modules": [
        {
            "path": [
                "root"
            ],
            "outputs": {
                "bucket_backups": {
                    "sensitive": false,
                    "type": "string",
                    "value": "tsam.backups"
                },
                "bucket_internal_backups": {
                    "sensitive": false,
                    "type": "string",
                    "value": "tsam.internal.backups"
                }
            },
            "resources": {
                "aws_s3_bucket.backups": {
                    "type": "aws_s3_bucket",
                    "depends_on": [],
                    "primary": {
                        "id": "tsam.backups",
                        "attributes": {
                            "acceleration_status": "",
                            "acl": "private",
                            "arn": "arn:aws:s3:::tsam.backups",
                            "bucket": "tsam.backups",
                            "bucket_domain_name": "tsam.backups.s3.amazonaws.com",
                            "force_destroy": "false",
                            "hosted_zone_id": "Z21DXDUVLTQW6Q",
                            "id": "tsam.backups",
                            "logging.#": "0",
                            "region": "eu-central-1",
                            "request_payer": "BucketOwner",
                            "tags.%": "0",
                            "versioning.#": "1",
                            "versioning.0.enabled": "false",
                            "versioning.0.mfa_delete": "false",
                            "website.#": "0"
                        },
                        "meta": {},
                        "tainted": false
                    },
                    "deposed": [],
                    "provider": "provider.aws"
                },
                "aws_s3_bucket.internal_backups": {
                    "type": "aws_s3_bucket",
                    "depends_on": [],
                    "primary": {
                        "id": "tsam.internal.backups",
                        "attributes": {
                            "acceleration_status": "",
                            "acl": "private",
                            "arn": "arn:aws:s3:::tsam.internal.backups",
                            "bucket": "tsam.internal.backups",
                            "bucket_domain_name": "tsam.internal.backups.s3.amazonaws.com",
                            "force_destroy": "false",
                            "hosted_zone_id": "Z21DXDUVLTQW6Q",
                            "id": "tsam.internal.backups",
                            "logging.#": "0",
                            "region": "eu-central-1",
                            "request_payer": "BucketOwner",
                            "tags.%": "0",
                            "versioning.#": "1",
                            "versioning.0.enabled": "true",
                            "versioning.0.mfa_delete": "false",
                            "website.#": "0"
                        },
                        "meta": {},
                        "tainted": false
                    },
                    "deposed": [],
                    "provider": "provider.aws"
                }
            },
            "depends_on": []
        }
    ]
}
`
