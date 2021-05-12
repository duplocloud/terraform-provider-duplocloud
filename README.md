# Terraform Provider duplocloud

## Releasing new versions

This repo now uses the `git flow` tool to perform releases.

There is a `scripts/release.sh` tool to automate the process.

You can learn more about this in the [README-release-process.md](README-release-process.md) doc.

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