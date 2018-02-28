#!/bin/bash -x

if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
    docker login azureclidev.azurecr.io -u $AZURESDKDEV_ACR_SP_USERNAME -p $AZURESDK_ACR_SP_PASSWORD
    echo 'Logged in to container registry'
    bash tools/build_container.sh
fi