if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
    docker login azureclidev.azurecr.io -u $AZURESDKDEV_ACR_SP_USERNAME -p $AZURESDK_ACR_SP_PASSWORD
    bash tools/build_container.sh
fi