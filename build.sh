#!/bin/bash

go mod init terraform-provider-duplocloud
go mod vendor
make build
#make release
make install


