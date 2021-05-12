# Release Procesa

## Performing a release

### Step 1 - release start

**NOTE: In the future, this might be moved to a github action.**

Run `scripts/release.sh start`.

What does this do?

  - This will run `git flow release start`, which will:
    - Checkout a new branch `release/MY.CURRENT.VERSION` from `develop` and push it to github
  - NOTE: If you forgot to "bump" the version after your prior release, an error will be given.

## Step 2 - release finish

**NOTE: In the future, this should be moved to a github action.**

Run `scripts/release.sh finish`.

What does this do?

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

### Step 3 - Publishing to github

This is taken care of automatically by the `.github/workflows/release.yml`

How does this work?

  - It uses the `goreleaser` tool to build and publish the release.
  - It is triggered by the release tag pushed by the "Finishing a release" step above.
  - Hashicorp has a webhook installed in the git repo which will then pull our release into the registry.
    - It is triggered automatically whenever a github release is created

### Step 4 - Bumping teh version

Run `scripts/release.sh next MY.NEW.VERSION`

  - This will:
    - Set the release version in the `Makefile` and all relevant example files.
    - Commit the changes and push to github.

