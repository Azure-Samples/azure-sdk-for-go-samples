GO ?= go
GH_ORG ?= Azure-Samples
ROOT = github.com/$(GH_ORG)/azure-sdk-for-go-samples
BASE = $(GOPATH)/src/$(ROOT)

# tested on 1.9+, no need to exclude /vendor
# to specify packages to skip: `PKGS_SKIP_RE='storage|graphrbac' make ...`
PKGS_SKIP_RE ?= ''
PKGS         != $(GO) list ./... | sed 's@$(ROOT)/@./@' | grep -v -E "$(PKGS_SKIP_RE)"

# uses Azure resources
test: dep
	$(GO) test -v $(PKGS)

test_pr: lint
	$(GO) build -v $(PKGS)

dep:
	$(GO) get -u github.com/golang/dep/cmd/dep
	cd $(BASE) && dep ensure

lint: dep
	$(GO) get -v github.com/alecthomas/gometalinter
	gometalinter --update
	# TODO: fix problems and enable all tests
	# TODO: address warnings
	gometalinter --errors \
		--enable=gofmt \
		--enable=goimports \
		--disable=vet \
		--disable=gotype \
		--disable=megacheck \
		$(PKGS)

.PHONY: test test_pr dep lint
