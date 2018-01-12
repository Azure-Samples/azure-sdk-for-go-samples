this script runs the samples, and cleans thse subscription
if [ "$TRAVIS_PULL_REQUEST" != "false" ]; then
    echo 'skip running tests'
    exit 0
fi

REALEXITSTATUS=0

dirs=$(go list ./... | grep -v /vendor/)
for d in $dirs
do
    go test -v $d
    REALEXITSTATUS=$(($REALEXITSTATUS+$?))
done

go install github.com/Azure-Samples/azure-sdk-for-go-samples/tools/cleanup
REALEXITSTATUS=$(($REALEXITSTATUS+$?))

cleanup -quiet

exit $REALEXITSTATUS
