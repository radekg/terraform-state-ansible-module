package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tfConfig "github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
)

var (
	terraformConfigUnderTest     string
	terraformLocalStateUnderTest string
	terraformVarsUnderTest       string
	argsTerraformTfUnderTest     string
	argsVarsTfUnderTest          string
)

func TestMain(m *testing.M) {
	tempDir, _ := ioutil.TempDir("", "tsam-test")

	terraformLocalStateUnderTest := filepath.Join(tempDir, "tempStateFile.tfstate")
	ioutil.WriteFile(terraformLocalStateUnderTest, []byte(testState), 0644)

	terraformConfigUnderTest = filepath.Join(tempDir, "terraform.tf")
	ioutil.WriteFile(terraformConfigUnderTest, []byte(fmt.Sprintf(testTerraformConfig, terraformLocalStateUnderTest)), 0644)

	terraformVarsUnderTest = filepath.Join(tempDir, "vars.tf")
	ioutil.WriteFile(terraformVarsUnderTest, []byte(testVars), 0644)

	argsTerraformTfUnderTest = filepath.Join(tempDir, "terraformTfArgs")
	ioutil.WriteFile(argsTerraformTfUnderTest, []byte(fmt.Sprintf(testTerrafromTfArgs, terraformConfigUnderTest)), 0644)

	argsVarsTfUnderTest = filepath.Join(tempDir, "varsTfArgs")
	ioutil.WriteFile(argsVarsTfUnderTest, []byte(fmt.Sprintf(testVarsTfArgs, terraformVarsUnderTest)), 0644)

	result := m.Run()
	os.RemoveAll(tempDir)
	os.Exit(result)
}

/*func TestTerraformTfExecution(t *testing.T) {
	executeProgram([]string{"program-name", argsTerraformTfUnderTest})
}*/

/*func TestVarsTfExecution(t *testing.T) {
	executeProgram([]string{"program-name", argsVarsTfUnderTest})
}*/

func TestValidRetrieve(t *testing.T) {
	ma := ModuleArgs{}
	var err error
	err = ma.validateRetrieve("o/output_name")
	if err != nil {
		t.Fatal("Expected no error o/output_name")
	}
	err = ma.validateRetrieve("r/resource/attribute")
	if err != nil {
		t.Fatal("Expected no error for r/resource/attribute")
	}
}

func TestFailsWithInvalidRetrieve(t *testing.T) {
	ma := ModuleArgs{}
	var err error
	err = ma.validateRetrieve("r/resource.only")
	if err == nil {
		t.Fatal("Expected the validation to fail")
	}
	err = ma.validateRetrieve("r/resource/attribute/wrong")
	if err == nil {
		t.Fatal("Expected the validation to fail")
	}
	err = ma.validateRetrieve("u/unknown/resource")
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
		TerraformFilePath: "/not/important/here.tf",
		State:             defaultState,
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
		TerraformFilePath: "/not/important/here.tf",
		State:             defaultState,
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

func TestFailUnsupportedTerraformFile(t *testing.T) {
	margs := &ModuleArgs{
		TerraformFilePath: "/not/important/outputs.tf",
		State:             defaultState,
	}
	_, err := margs.validate()
	if err == nil {
		t.Fatalf("Expected an error")
	}
}

func TestHandlingCorrectVariables(t *testing.T) {
	vars := []*tfConfig.Variable{
		&tfConfig.Variable{
			Name:    "test_string_variable",
			Default: "string value",
		},
		&tfConfig.Variable{
			Name:    "test_list_variable",
			Default: []string{"string value", "string value2"},
		},
		&tfConfig.Variable{
			Name: "test_map_variable",
			Default: map[string]interface{}{
				"hello": "test world!",
			},
		},
	}

	responseData, err := processVariables(vars)
	if err != nil {
		t.Fatalf("Expected no error")
	}
	if m, ok := responseData["test_string_variable"]; !ok {
		t.Fatalf("Expected test_string_variable in response data")
		if _, ook := m.(map[string]interface{})["default"]; !ook {
			t.Fatalf("Expected test_string_variable.default in response data")
		}
	}
	if m, ok := responseData["test_list_variable"]; !ok {
		t.Fatalf("Expected test_list_variable in response data")
		if _, ook := m.(map[string]interface{})["default"]; !ook {
			t.Fatalf("Expected test_list_variable.default in response data")
		}
	}
	if m, ok := responseData["test_map_variable"]; !ok {
		t.Fatalf("Expected test_map_variable in response data")
		if _, ook := m.(map[string]interface{})["default"]; !ook {
			t.Fatalf("Expected test_map_variable.default in response data")
		}
	}
}

func TestFailUnknownVariableType(t *testing.T) {
	vars := []*tfConfig.Variable{
		&tfConfig.Variable{
			Name:         "test_string_variable",
			Default:      "string value",
			DeclaredType: "unknown_type",
		},
	}
	_, err := processVariables(vars)
	if err == nil {
		t.Fatalf("Expected an error")
	}
}

var testTerrafromTfArgs = `{
	"terraform_file_path": "%s",
	"retrieves": [
		{ "retrieve": "o/bucket_backups" },
		{ "retrieve": "r/aws_s3_bucket.backups/bucket_domain_name" },
		{ "retrieve": "o/bucket_internal_backups" },
		{ "retrieve": "r/aws_s3_bucket.internal_backups/bucket_domain_name" },
		{ "retrieve": "r/aws_s3_bucket.internal_backups/request_payer" }
	]
}`

var testVarsTfArgs = `{
	"terraform_file_path": "%s"
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

var testVars = `
variable "key" {
	type = "string"
	default = "test value"
}
variable "images" {
	type = "map"
	default = {
		us-east-1 = "image-1234"
		us-west-2 = "image-4567"
	}
}
variable "zones" {
	default = ["us-east-1a", "us-east-1b"]
}
`
