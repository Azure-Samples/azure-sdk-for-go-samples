FROM golang:1.10

LABEL a01.product="azuresdkforgo"

COPY . /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure -v

RUN mkdir /app
RUN go run tools/list/list.go > /app/test_index