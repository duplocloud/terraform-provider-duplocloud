#!/bin/bash -eu
#
# This script starts or finishes a release using git flow

# Prints the current version
echo_version() {
    sed -ne 's/^VERSION=\([0-9\.]*\).*$/\1/p' <Makefile
}

# Starts a release using git flow
release_start() {
    local version
    version="$(echo_version)"

    GIT_MERGE_AUTOEDIT=no git flow release start "$version"
}

# Finishes a release using git flow
release_finish() {
    local version
    version="$(echo_version)"

    # Switch to the release branch.
    git checkout "release/$version"

    # Run tests
    make test

    # Prepare local branches
    git checkout develop ; git pull
    git checkout master ; git pull
    git checkout "release/$version"

    # Update any generated code
    make doc
    git add docs examples
    git commit -m 'update generated docs and examples' docs examples || true

    # Finish the release
    GIT_MERGE_AUTOEDIT=no git flow release finish "$version"

    # Push updated master and tag to github
    git checkout master ; git push
    git checkout "v$version" ; git push origin "v$version"

    # Push the updated develop to master
    git checkout develop ; git push
}

# Changes the version number in all relevant files
release_set_version() {
    local version
    version="$1"

    trap 'rm -f Makefile~ examples/*/main.tf~' EXIT
    
    sed -e 's/^\(VERSION=\)[0-9\.]*/\1'"$version"'/g' -i~ Makefile
    find examples -name main.tf -exec \
        sed -e 's/\(version = "\)[0-9\.]*\(".*# RELEASE VERSION\)/\1'"$version"'\2/g' -i~ \{\} \;
}

action="$1" ; shift
case "$action" in
start)
    release_start "$@"
    ;;
finish)
    release_finish "$@"
    ;;
next)
    release_set_version "$1"
    git add Makefile examples/*/main.tf examples/emr/files/*/main.tf
    GIT_MERGE_AUTOEDIT=no git commit -m '[release] bump version to '"$1"
    git push origin develop:develop
    ;;
set-version)
    release_set_version "$1"
    ;;
*)
    echo "Usage: release.sh {start|finish|next}"
    exit 1
esac
