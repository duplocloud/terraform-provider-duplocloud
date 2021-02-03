#!/bin/bash

duploObjectSnakecase=$1;
duploObjectCamelcase=$2;

#duploObjectCamelcase = DuploService
#duploObjectSnakecase = duplo_service

if [ -z "$duploObjectCamelcase" ]
then
      echo "arguments required 1=name of file, 2=name of object "
      echo "eg. 1=aws_redis_cache; 2=AwsRedisCache "
else
      echo " creating $duploObjectSnakecase $duploObjectCamelcase"
fi

echo " creating $duploObjectSnakecase $duploObjectCamelcase "

#foldeers
export sdk_path=../duplosdk
export duplocloud_path=../duplocloud

model_file="$sdk_path/${duploObjectSnakecase}.go"
resource_file="$duplocloud_path/resource_duplo_${duploObjectSnakecase}.go"

echo " creating '$model_file' '$resource_file' "

#create files
cp model_xvyzw.go                          "$model_file"
cp resource_xvyzw.go                       "$resource_file"

# regex
# --find replace Xvyzw to DuploService
sed  -i  ""  "s/Xvyzw/$duploObjectCamelcase/g"    "$model_file"
sed  -i  ""  "s/Xvyzw/$duploObjectCamelcase/g"    "$resource_file"

echo " created $duploObjectSnakecase $duploObjectCamelcase => '$model_file' '$resource_file'  "
