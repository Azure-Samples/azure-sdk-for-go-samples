#!/bin/bash -x

image_owner=${TRAVIS_REPO_SLUG%-Samples/azure-sdk-for-go-samples} 
image_owner=${image_owner:="private-${USER}"}
image_owner=`echo $image_owner | tr '[:upper:]' '[:lower:]'`

build=$TRAVIS_BUILD_NUMBER
build=${build:=`date +%Y%m%d%H%M`}

EXIT_CODE=0

mv tools/Dockerfile .

echo 'Building docker image'
docker build -t azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-$build .
EXIT_CODE=$(($EXIT_CODE+$?))

echo 'Pushing docker image'
docker push azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-$build
EXIT_CODE=$(($EXIT_CODE+$?))

mv Dockerfile tools

exit $EXIT_CODE