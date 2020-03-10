__exitcode=0
for package in $packages; do
    echo calling: go test -v -timeout 12h $(convert_path_to_package $package)
    go test -v -timeout 12h $(convert_path_to_package $package)
    if [ $? -ne 0 ]; then __exitcode=1; fi
    echo
done

exit $__exitcode