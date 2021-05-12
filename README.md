# Terraform Provider duplocloud

## Building locally.

 - To build and install, run `make install`.
 - To generate or update documentation, run `make doc`.

## Running examples.

```
cd examples/service
terraform init && terraform apply
```

## Using debug output during execution.

``` 
export TF_LOG_PATH=duplo.log
export TF_LOG=TRACE
terraform init && terraform apply
```