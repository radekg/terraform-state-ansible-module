package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var terraformConfigUnderTest string

func TestMain(m *testing.M) {

	tempStateFile, _ := ioutil.TempFile("", "tempStateFile*.tfstate")
	ioutil.WriteFile(tempStateFile.Name(), []byte(testState), 0644)

	tempTerraformConfigFile, _ := ioutil.TempFile("", "tempTerraformConfigFile*.tf")
	ioutil.WriteFile(tempTerraformConfigFile.Name(), []byte(fmt.Sprintf(testTerraformConfig, tempStateFile.Name())), 0644)

	tempArgsFile, _ := ioutil.TempFile("", "tempArgsFile")
	ioutil.WriteFile(tempArgsFile.Name(), []byte(fmt.Sprintf(testArgs, tempTerraformConfigFile.Name())), 0644)
	terraformConfigUnderTest = tempArgsFile.Name()

	defer os.Remove(tempStateFile.Name())
	defer os.Remove(tempTerraformConfigFile.Name())
	defer os.Remove(tempArgsFile.Name())

	result := m.Run()

	os.Exit(result)

}

func TestExecution(t *testing.T) {
	executeProgram([]string{"program-name", terraformConfigUnderTest})
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
