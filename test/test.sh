#!/usr/bin/env bash
## standard helpers
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
__exitcode=0 # on any errors will be set to 1
source "${__dirname}/_functions.sh"

echo ISPR: $ISPR

echo installing dependencies...
install_dependencies $ROOTDIR
if [ $? -ne 0 ]; then __exitcode=1; fi
echo installed dependencies.
echo

echo list of packages to test:
packages=$(list_packages_to_test "$ROOTDIR" "$TESTDIR/excluded_packages")
for package in $packages; do
    echo $package | sed "s@$ROOTDIR/\(.*\)@\t\1@"
done
echo

echo packages excluded by test/excluded_packages:
for package in $(cat "${__dirname}/excluded_packages"); do
    echo $package | sed "s@\(.*\)@\t\1@"
done
echo

for package in $packages; do
    echo testing package $package
    echo linting...
    run_lint_on_packages $package
    if [ $? -ne 0 ]; then __exitcode=1; fi
    echo done linting.
    echo building...
    build_packages $(convert_path_to_package $package)
    if [ $? -ne 0 ]; then __exitcode=1; fi
    echo done building.
    echo
done

if [ "x$TRAVIS_PULL_REQUEST" != "x" -a "x$TRAVIS_PULL_REQUEST" != "xfalse" ]; then
    ISPR=1
fi

# don't run live tests on PRs
echo ISPR: $ISPR
if [ "x$ISPR" = "x" -o "x$ISPR" = "x0" ]; then
    for package in $packages; do
        echo calling: go test -v -timeout 12h $(convert_path_to_package $package)
        go test -v -timeout 12h $(convert_path_to_package $package)
        if [ $? -ne 0 ]; then __exitcode=1; fi
        echo
    done
else
    echo "live tests are skipped on PRs"
    __exitcode=0
fi

# indicate failure if needed
exit $__exitcode
