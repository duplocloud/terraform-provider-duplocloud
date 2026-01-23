# Terraform Provider duplocloud

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Table of Contents
- [Overview](#overview)
- [Release Process](#releasing-new-versions)
- [Development](#development)
  - [Building Locally](#building-locally)
  - [Running Examples](#running-examples)
  - [Debugging](#debugging)
- [WSL2 Setup](#installing-and-running-project-on-wsl2)
- [License](#license)

## Overview
The DuploCloud Terraform Provider enables infrastructure management through DuploCloud's platform. It provides resources and data sources for interacting with the DuploCloud API.

## Releasing new versions

This repository uses GitHub Actions for automated releases. We support two types of releases:
- **Hotfix Release**: For urgent production fixes
- **Major Release**: For planned feature releases

### Useful Links
- [Detailed Release Process Documentation](release-process.md)
- [GitHub Actions Workflows](.github/workflows)
  - [Start Release](.github/workflows/start-release.yml)
  - [Finish Release](.github/workflows/finish-release.yml)
  - [Start Hotfix](.github/workflows/start-hotfix.yml)
  - [Finish Hotfix](.github/workflows/finish-hotfix.yml)
  - [Release Publish](.github/workflows/release-publish.yml)
- [Architecture Documentation](ARCHITECTURE.md)
- [Security Policy](SECURITY.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)

For detailed information about the release process, please refer to our [Release Process Documentation](release-process.md).

## Development

### Building Locally

 - To build and install, run `make install`.
 - To generate or update documentation, run `make doc`.

### Running Examples

```shell
cd examples/service
terraform init && terraform apply
```

### Debugging

#### Using debug output during execution

``` shell
export TF_LOG_PATH=duplo.log
export TF_LOG=TRACE
terraform init && terraform apply
```

#### Using a debugger on VS code

Reference documentation [here](https://developer.hashicorp.com/terraform/plugin/debugging#debugger-based-debugging)

1. Install the VS code [GO extension](https://marketplace.visualstudio.com/items?itemName=golang.go)
2. Go to VS code debug tab and start a new debug session for the terraform provider. You should see something similar to the output below in the console:
```
Starting: /Users/{USERNAME}/go/bin/dlv dap --listen=127.0.0.1:55046 --log-dest=3 from /Users/{USERNAME}/Desktop/duplocloud/terraform-provider-duplocloud
DAP server listening at: 127.0.0.1:55046
Type 'dlv help' for list of commands.
Provider started. To attach Terraform CLI, set the TF_REATTACH_PROVIDERS environment variable with the following:
	TF_REATTACH_PROVIDERS='{"registry.terraform.io/duplocloud/duplocloud":{"Protocol":"grpc","ProtocolVersion":5,"Pid":79817,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/32/bgvj1ynd6p16109l4t7sx4jm0000gn/T/plugin2604224326"}}}'
```
3. Copy `TF_REATTACH_PROVIDERS` to your clipboard and a terminal session to run the terraform commands
4. Run the terraform apply command like the following:
```
TF_REATTACH_PROVIDERS='{"registry.terraform.io/duplocloud/duplocloud":{"Protocol":"grpc","ProtocolVersion":5,"Pid":79817,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/32/bgvj1ynd6p16109l4t7sx4jm0000gn/T/plugin2604224326"}}}' terraform apply
```
5. Happy debugging!

## Installing and Running Project on WSL2
*Assumptions*
1. Wsl2 is already installed
2. You are using a Ubuntu Linux
3. You have pulled the project on Wsl2
4. You will need to enable systemd support https://devblogs.microsoft.com/commandline/systemd-support-is-now-available-in-wsl/#set-the-systemd-flag-set-in-your-wsl-distro-settings

Install make on Wsl2:

```shell
sudo apt-get install build-essential
make version # make: *** No rule to make target 'version'.  Stop.
```

Install go on Wsl2:
```shell
sudo apt install golang-go
go version # Test your install: you should see something like "go version go1.18.1 linux/amd64"
```

### Build Project
Make sure you are in the directory where the **Makefile** is located:
```shell
make install
```

### Install Terraform
```shell
sudo snap install terraform --classic
terraform version # you should see something like Terraform v1.6.5 ....
```

## License
This project is licensed under the [MIT License](LICENSE).