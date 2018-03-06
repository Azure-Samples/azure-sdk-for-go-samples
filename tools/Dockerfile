FROM golang:1.10-alpine

COPY . /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
ADD https://a01tools.blob.core.windows.net/droid/0.19.0/linux/a01droid /app/a01droid
ADD https://a01tools.blob.core.windows.net/droid/0.19.0/linux/a01dispatcher /app/a01dispatcher

RUN apk add --no-cache git bash && \
    go get github.com/golang/dep/cmd/dep && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/get_index /app/get_index && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/metadata.yml /app/metadata.yml && \
    chmod +x /app/a01droid && \
    chmod +x /app/a01dispatcher && \
    chmod +x /app/get_index && \
    go run /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/list/list.go > /app/test_index


WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure

WORKDIR /
CMD /app/a01droid