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

    GIT_MERGE_AUTOEDIT=no git flow release finish "$version"
    git push origin develop:develop
    git push origin "$version"
    (git checkout master && git pull && git push origin master:master)
    git checkout develop
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
    git add Makefile examples/*/main.tf
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
