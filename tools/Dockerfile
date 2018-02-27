FROM golang:1.10-alpine

LABEL a01.product="azuresdkforgo" \
    a01.index.schema="v2" \
    a01.env.AZ_SUBSCRIPTION_ID="secret:sp.subscriptionid" \
    a01.env.AZ_TENANT_ID="secret:sp.tenant" \
    a01.env.AZ_CLIENT_ID="secret:sp.clientid" \
    a01.env.AZ_CLIENT_SECRET="secret:sp.clientsecret" \
    a01.env.AZ_SP_OBJECT_ID="secret:sp.objectid" \
    a01.env.AZ_LOCATION="arg-mode"

COPY . /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
ADD https://a01tools.blob.core.windows.net/droid/latest/linux/a01droid /app/a01droid

RUN apk add --no-cache git bash && \
    go get github.com/golang/dep/cmd/dep && \
    chmod +x /app/a01droid && \
    mv /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/get_index /app/get_index && \
    go run /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples/tools/list/list.go > /app/test_index


WORKDIR /go/src/github.com/Azure-Samples/azure-sdk-for-go-samples
RUN dep ensure

WORKDIR /
CMD /app/a01droid