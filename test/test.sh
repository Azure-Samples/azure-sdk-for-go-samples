#!/usr/bin/env bash
## standard helpers
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
__exitcode=0 # on any errors will be set to 1
source "${__dirname}/_functions.sh"

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
    echo building...
    build_packages $(convert_path_to_package $package)
    if [ $? -ne 0 ]; then __exitcode=1; fi
    echo done building.
    echo
done
