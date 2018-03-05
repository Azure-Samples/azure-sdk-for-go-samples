dirs=$(go list ./...)

EXIT_CODE=0
for d in $dirs; do
    go test -i $d 2> output;
    grep -v 'vendor' output > novendoroutput; # there are some build errors on dependencies
    cat novendoroutput;
    if [ -s novendoroutput ]; then
        EXIT_CODE=$(($EXIT_CODE+1));
    fi;
done

rm output
rm novendoroutput

exit $EXIT_CODE