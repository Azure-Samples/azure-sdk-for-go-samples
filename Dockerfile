FROM golang:1.10-alpine

LABEL a01.product="azuresdkforgo"
LABEL a01.index.schema="v2"
LABEL a01.env.AZURE_SUBSCRIPTION_ID="secret:sp.go.subscriptionid"
LABEL a01.env.AZURE_TENANT_ID="secret:sp.tenant"
LABEL a01.env.AZURE_CLIENT_ID="secret:sp.go.clientid"
LABEL a01.env.AZURE_CLIENT_SECRET="secret:sp.go.clientsecret"

# install additional stuff alpine does not have
RUN apk add --no-cache git bash

# get everything needed to build and run tests
COPY . /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure

# add get_index and droid
WORKDIR /
RUN mkdir /app
RUN mv go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/get_index /app/get_index
RUN mv go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/a01droid /app/a01droid
RUN go run /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/list/list.go > /app/test_index

CMD app/a01droid