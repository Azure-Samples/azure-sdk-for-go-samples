# this script runs the samples, and cleans thse subscription
if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
    exit 0
fi

REALEXITSTATUS=0

dirs=$(go list ./...)
test -z "`for d in $dirs; do go test -v $d | tee /dev/stderr; done`"
REALEXITSTATUS=$(($REALEXITSTATUS+$?))

go install github.com/Azure-Samples/azure-sdk-for-go-samples/tools/cleanup
REALEXITSTATUS=$(($REALEXITSTATUS+$?))

cleanup -quiet
REALEXITSTATUS=$(($REALEXITSTATUS+$?))

exit $REALEXITSTATUS
