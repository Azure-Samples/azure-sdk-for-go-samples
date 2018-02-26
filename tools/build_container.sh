curl -sL https://a01tools.blob.core.windows.net/droid/linux/a01droid -o tools/a01droid
chmod +x tools/a01droid
chmod +x tools/get_index

image_owner=${TRAVIS_REPO_SLUG%-Samples/azure-sdk-for-go-samples} 
image_owner=${image_owner:="private-${USER}"}
image_owner=`echo $image_owner | tr '[:upper:]' '[:lower:]'`

build=$TRAVIS_BUILD_NUMBER
build=${build:=`date +%Y%m%d%H%M`}

mv tools/dockerfiles/* .

docker build -t azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-prod-$build -f Dockerfile.1.10.prod .
docker push azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-prod-$build
docker build -t azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-canary-$build -f Dockerfile.1.10.canary .
docker push azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-canary-$build

mv `ls | grep ^Dockerfile` tools/dockerfiles