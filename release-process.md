# Release Process Documentation

This document outlines the two main release processes for the Terraform Provider DuploCloud:
1. Hotfix Release - for urgent production fixes
2. Major Release - for planned feature releases and updates

## 1. Hotfix Release Process

**Workflows Used:**
- Start Hotfix (`start-hotfix.yml`)
- Finish Hotfix (`finish-hotfix.yml`)
- Release Publish (`release-publish.yml`)

Hotfix releases are used for urgent fixes to the production codebase. The process uses dedicated workflows to automate branch creation, version bumping, and merging.

**Hotfix Flow:**

    [Manual Trigger: Start Hotfix]
            |
    [Checkout master]
            |
    [Determine Hotfix Version]
            |
    [Setup Git Config]
            |
    [Start gitflow hotfix branch]
            |
    [Update Docs & Bump Version]
            |
    [Open PR to master]
            |
    [PR Merged to master triggers Finish Hotfix]
            |
    [Checkout master]
            |
    [Setup Git Config]
            |
    [Finish gitflow hotfix branch]
            |
    [Tag triggers Publish Release]
            |
    [Build, Test, Draft Release]
            |
    [Publish Release]

## 2. Major Release Process

**Workflows Used:**
- Start Release (`start-release.yml`)
- Finish Release (`finish-release.yml`)
- Release Publish (`release-publish.yml`)

Major releases are for new features and planned updates. The process ensures versioning, documentation, and release publication are all handled automatically.

**Major Release Flow:**

    [Manual Trigger: Start Release]
            |
    [Checkout develop]
            |
    [Determine Release Version]
            |
    [Setup Git Config]
            |
    [Start gitflow release branch]
            |
    [Update Docs & Bump Version]
            |
    [Open PR to master]
            |
    [PR Merged to master triggers Finish Release]
            |
    [Checkout master]
            |
    [Setup Git Config]
            |
    [Finish gitflow release branch]
            |
    [Bump version on develop]
            |
    [Commit & Push to develop]
            |
    [Tag triggers Publish Release]
            |
    [Build, Test, Draft Release]
            |
    [Publish Release]

## Process Details

### Hotfix Process Details
1. **Start Hotfix:**
   - Triggered manually via GitHub Actions
   - Creates branch from `master`
   - Auto-increments patch version
   - Updates documentation and version numbers
   - Opens PR to `master`

2. **Finish Hotfix:**
   - Triggered when PR is merged to `master`
   - Completes the hotfix using gitflow
   - Tags the release

3. **Publish Release:**
   - Triggered by tag push
   - Runs tests
   - Builds and signs artifacts
   - Creates GitHub release
   - Publishes to registry

### Major Release Details
1. **Start Release:**
   - Triggered manually via GitHub Actions
   - Creates branch from `develop`
   - Uses version from Makefile or override
   - Updates documentation and version numbers
   - Opens PR to `master`

2. **Finish Release:**
   - Triggered when PR is merged to `master`
   - Completes the release using gitflow
   - Bumps version on `develop`
   - Creates tag

3. **Publish Release:**
   - Triggered by tag push
   - Runs tests
   - Builds and signs artifacts
   - Creates GitHub release
   - Publishes to registry

## Best Practices

1. **For Hotfixes:**
   - Use only for urgent production fixes
   - Always branch from `master`
   - Keep changes minimal and focused
   - Test thoroughly before merging

2. **For Major Releases:**
   - Ensure all features are in `develop`
   - Review documentation updates
   - Run full test suite
   - Plan release timing with team

## Troubleshooting

- If workflows fail, check GitHub Actions logs
- Ensure all required secrets are configured
- Verify GPG keys are valid
- Check branch protection rules
- Confirm proper access permissions
