FROM golang:1.10-alpine

COPY . /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples

RUN apk add --no-cache git bash && \
    go get github.com/golang/dep/cmd/dep && \
    mkdir app  && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/get_index /app/get_index && \
    chmod +x /app/get_index && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/metadata.yml /app/metadata.yml && \
    go run /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/list/list.go > /app/test_index


WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure

WORKDIR /
CMD /mnt/agents/a01droid