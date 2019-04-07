# Ansible module for retrieving Terraform state properties

[![CircleCI](https://circleci.com/gh/radekg/terraform-state-ansible-module.svg?style=svg)](https://circleci.com/gh/radekg/terraform-state-ansible-module)

Coming full circle: provision machines [using Terraform and Ansible](https://github.com/radekg/terraform-provisioner-ansible), then use the Terraform state to power your Ansible!

## Installation

Download latest release from the [prebuilt releases on GitHub](https://github.com/radekg/terraform-state-ansible-module/releases). Rename the released binary to `terraform_state` and place it in `~/.ansible/plugins/modules` or any other valid Ansible module directory.

## Usage

The module supports two Terraform file types: `terraform.tf` and `vars.tf`. When passing the path to `terraform.tf` file, an existence of a properly configured `backend` is assumed.

### Handling state using `terraform.tf`

```yaml
---
- hosts: all
  become: no
  connection: local
  gather_facts: no
  tasks:
    - terraform_state:
        terraform_file_path: /path/to/my/terraform.tf
        retrieves:
         - { retrieve: o/bucket_backups }
         - { retrieve: r/aws_s3_bucket.backups/bucket_domain_name }
         - { retrieve: r/aws_s3_bucket.backups/bucket }
      register: retrieved_data
    - set_fact:
        retrieved_terraform_state_data: "{{ retrieved_data.msg | from_json }}"
    - debug:
        msg: " ==> I have got {{ retrieved_terraform_state_data.aws_s3_bucket.backups }}"
```

### Handling variables using `vars.tf`

```yaml
---
- hosts: all
  become: no
  connection: local
  gather_facts: no
  tasks:
    - terraform_state:
        terraform_file_path: /path/to/my/vars.tf
      register: retrieved_vars
    - set_fact:
        retrieved_vars_data: "{{ retrieved_vars.msg | from_json }}"
    - debug:
        msg: " ==> I have these vars {{ retrieved_vars_data }}"
```

The data retrieved from the `vars.tf` file contains only the values provided with the `default` property. Furthermore, the `default` property existence is explicit in the output. For a `vars.tf` file:

```tf
variable "key" {
    default = "test value"
}
```

the value is avialble via `<retrieved_vars>.key.default`.

## Parameters

- `terraform_file_path`: string, required, path to the `terraform.tf` or `vars.tf` file, no default

Following configuration settings take effect only when `terraform.tf` is used:

- `state`: string, optional, state name to use, default value `default`
- `retrieves`: a list of values to retrieve, at least one is required, each retrieve
  - `retrieve`: string, required, property to retrieve; more below
  - `module_path`: string, optional, default `empty string` (maps to the `root` path)
- `require_all`: boolean, default `false`; if `true`, the module will fail when at least one of the values could not be found in the Terraform state

### Retrieve paths

Currently, only two possible combinations can be reetrieved:

- output: `o/<output name>`
- resource: `r/<resource>/<attribute>`

## Creating releases

To cut a release, run: 

    curl -sL https://raw.githubusercontent.com/radekg/git-release/master/git-release --output /tmp/git-release
    chmod +x /tmp/git-release
    /tmp/git-release --repository-path=$GOPATH/src/github.com/radekg/terraform-state-ansible-module
    rm -rf /tmp/git-release

After the release is cut, build the binaries for the release:

    git checkout v${RELEASE_VERSION}
    ./bin/build-release-binaries.sh

After the binaries are built, upload the to GitHub release.
