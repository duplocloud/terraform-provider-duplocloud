# Terraform Provider duplocloud

## Releasing new versions

This repo now the Duplo Github Actions to perform releases.

There is a `Start Release` workflow to kick off the process.

## Development

### Building locally.

 - To build and install, run `make install`.
 - To generate or update documentation, run `make doc`.

### Running examples.

```
cd examples/service
terraform init && terraform apply
```

### Using debug output during execution.

``` 
export TF_LOG_PATH=duplo.log
export TF_LOG=TRACE
terraform init && terraform apply
```
