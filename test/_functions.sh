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

# run_lint_on_packages runs chosen linters on specified packages
function run_lint_on_packages {
    _packages=$1

	  golangci-lint run ${_packages}
}

# build_packages build packages
function build_packages {
    _packages=$1
    echo $_packages
    go build -v ${_packages}
}

# run_test_on_packages runs `go test -v` on specified packages
# long timeout so tests can finish
function run_test_on_packages {
    _packages=$1
    go test -timeout 12h -v ${_packages}
}
