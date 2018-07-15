test:
	./test/test.sh

testpr:
	ISPR=1 ./test/test.sh

.PHONY: test testpr
