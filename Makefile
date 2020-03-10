test:
	./test/test.sh

testcommit:
    ./test/test_on_commit.sh

.PHONY: test testcommit
