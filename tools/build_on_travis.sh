#!/bin/bash -x
echo $TRAVIS_PULL_REQUEST
echo $TRAVIS_BRANCH
echo $TRAVIS_GO_VERSION

if [ $TRAVIS_PULL_REQUEST != 'false' ] && [ $TRAVIS_BRANCH == 'master' ] && [ $TRAVIS_GO_VERSION = 1.10.*]; then
    echo 'Login to docker'
    docker login azureclidev.azurecr.io -p $AZURESDK_ACR_SP_PASSWORD -u $AZURESDKDEV_ACR_SP_USERNAME
    bash tools/build_container.sh
fi