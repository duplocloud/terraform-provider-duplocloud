# Terraform Provider duplocloud

## Releasing new versions

This repo now the Duplo Github Actions to perform releases.

There is a `Start Release` workflow to kick off the process.

## Development

### Building locally.

 - To build and install, run `make install`.
 - To generate or update documentation, run `make doc`.

### Running examples.

```shell
cd examples/service
terraform init && terraform apply
```

### Using debug output during execution.

``` shell
export TF_LOG_PATH=duplo.log
export TF_LOG=TRACE
terraform init && terraform apply
```

## Installing and running project on WSL2
*Assumptions*
1. Wsl2 is already installed
2. You are using a Ubuntu Linux
3. You have pulled the project on Wsl2

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

Build project
- Make sure you are in the directory where the **Makefile** is located
```
make install
```

Install terraform
```shell
sudo snap install terraform --classic
terraform version # you should see something like Terraform v1.6.5 ....
```