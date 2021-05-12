# Terraform Provider duplocloud

## Releasing new versions

This repo now uses the `git flow` tool to perform releases.

There is a `scripts/release.sh` tool to automate the process.

### Starting a release.

**NOTE: In the future, this might be moved to a github action.**

 - Run `scripts/release.sh start`.
   - This will run `git flow release start`, which will:
     - Checkout a new branch `release/MY.CURRENT.VERSION` from `develop` and push it to github
   - NOTE: If you forgot to "bump" the version after your prior release, an error will be given.

### Finishing a release.

**NOTE: In the future, this should be moved to a github action.**

 - Run `scripts/release.sh finish`.
   - This will:
     - Checkout the `release/MY.CURRENT.VERSION` branch
     - Run the unit tests
     - Prepare your local branch copies for finishing the release
     - Generate new documenation and commit it to git
     - Run `git flow release finish`, which will:
       - Prompt you for a tag message.
         - *PLEASE* enter a good description for the release here.
       - Merge `release/MY.CURRENT.VERSION` to `master`
       - Back-merge `master` into `develop`
       - Tag the release as `vMY.CURRENT.VERSION`
     - Push `develop`, `master` and the new release tag to github 

### Bumping the version after a release.

 - Run `scripts/release.sh next MY.NEW.VERSION`
   - This will:
     - Set the release version in the `Makefile` and all relevant example files.
     - Commit the changes and push to github.

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