FROM golang:1.10.3-stretch
ARG PACKAGE=github.com/Azure-Samples/azure-sdk-for-go-samples/apps/basic_web_app

RUN mkdir /app # for built artifacts
RUN mkdir -p $GOPATH/src/${PACKAGE}
WORKDIR $GOPATH/src/${PACKAGE}

# better for build cache to specify necessary files
ADD ["Gopkg.*","main.go", "./"]
RUN go get -u -v github.com/golang/dep/cmd/dep && dep ensure -v
RUN go build -v -o /app/server .

CMD ["/app/server"]
