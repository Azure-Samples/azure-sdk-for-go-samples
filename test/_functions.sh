TESTDIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOTDIR=$(cd "$TESTDIR/.." && pwd)
ROOTPKG="$ROOTDIR" | sed "s@$GOPATH@@"

# list_packages_to_test returns all go packages within ${base_path}
# except those listed in ${exclude_file}.
# example: `list_packages_to_test . ./test/excluded_packages`
function list_packages_to_test {
    base_path=$1
    exclude_file=$2

    base_package=$(echo $base_path | sed "s@$GOPATH/src/@@")

    excluded=$(cat ${exclude_file})
    packages=$(go list ${base_package}/... | grep -v -E "$excluded")

    for package in $packages; do
        echo $package | sed "s@^@$GOPATH/src/@"
    done
}

function convert_path_to_package {
    path=$1

    echo $path | sed "s@$GOPATH/src/@@"
}


# install_dependencies installs dependencies
function install_dependencies {
    base_path=$1

    go get -u github.com/golang/dep/cmd/dep
    cd $base_path && dep ensure -v
}

# run_lint_on_packages runs chosen linters on specified packages
function run_lint_on_packages {
    _packages=$1

	go get -u github.com/alecthomas/gometalinter
	gometalinter --install > /dev/null
	# TODO: fix problems and enable all tests
	# TODO: address warnings
	gometalinter --errors \
		--enable=gofmt \
		--enable=goimports \
		--disable=gotype \
		--disable=megacheck \
        --disable=vet \
		${_packages}
}

# build_packages build packages
function build_packages {
    _packages=$1
    echo $_packages
    go build -v ${_packages}
}

# run_test_on_packages runs `go test -v` on specified packages
function run_test_on_packages {
    _packages=$1
    go test -v ${_packages}
}
