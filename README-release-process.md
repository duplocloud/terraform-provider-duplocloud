# Release Process

## Quick List of Steps

There are four manual steps, one automatic step.

  - Run: `scripts/release.sh start`
  - Run: `scripts/release.sh finish`
  - Automatically Done: testing, building, archiving, publishing draft release.
  - Run: `scripts/release.sh next MY.NEXT.VERSION`
  - From Github UI:  Go to the release, and change from a draft to a published releasee.


## Internal details

You can learn more about this in the [README-release-process-internals.md](README-release-process-internals.md) doc.
