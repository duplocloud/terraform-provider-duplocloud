#!/bin/bash

AWS_SDK_VERSION=2.15.40
ICEBERG_VERSION=0.11.1
MAVEN_URL=https://repo1.maven.org/maven2
ICEBERG_MAVEN_URL=$MAVEN_URL/org/apache/iceberg
AWS_MAVEN_URL=$MAVEN_URL/software/amazon/awssdk
# NOTE: this is just an example shared class path between Spark and Flink,
#  please choose a proper class path for production.
LIB_PATH=/usr/share/aws/aws-java-sdk/

AWS_PACKAGES=(
  "bundle"
  "url-connection-client"
)

ICEBERG_PACKAGES=(
  "iceberg-spark3-runtime" 
)

install_dependencies () {
  install_path=$1
  download_url=$2
  version=$3
  shift
  pkgs=("$@")
  for pkg in "${pkgs[@]}"; do
    sudo wget -P $install_path $download_url/$pkg/$version/$pkg-$version.jar
  done
}

install_dependencies $LIB_PATH $ICEBERG_MAVEN_URL $ICEBERG_VERSION "${ICEBERG_PACKAGES[@]}"
install_dependencies $LIB_PATH $AWS_MAVEN_URL $AWS_SDK_VERSION "${AWS_PACKAGES[@]}"
