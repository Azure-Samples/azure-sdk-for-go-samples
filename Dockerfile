FROM golang:1.10-alpine

RUN mkdir -p /usr/local/bin/node
RUN apk add --no-cache git bash
RUN apk add --update nodejs nodejs-npm
RUN npm install -g n
RUN n 8.9.0
RUN git clone -b for-demo https://github.com/umar-muneer/azure-sdk-for-go-samples.git /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN go get github.com/golang/dep/cmd/dep && \
    mkdir /app && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/get_index /app/get_index && \
    chmod +x /app/get_index && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/metadata.yml /app/metadata.yml && \
    go run /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/list/list.go > /app/test_index
WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure
RUN sh install.sh
ENV AZURE_AUTH_LOCATION /mnt/secrets/authfile.json

WORKDIR /