image_owner=${TRAVIS_REPO_SLUG%-Samples/azure-sdk-for-go-samples} 
image_owner=${image_owner:="private-${USER}"}
image_owner=`echo $image_owner | tr '[:upper:]' '[:lower:]'`

build=$TRAVIS_BUILD_NUMBER
build=${build:=`date +%Y%m%d%H%M`}

mv tools/Dockerfile .

docker build -t azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-$build .
docker push azureclidev.azurecr.io/azuresdk-test-$image_owner:go1.10-$build

mv Dockerfile tools