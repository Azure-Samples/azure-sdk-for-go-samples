This repo provides samples to help developers understand how to interact with Azure services using the Azure SDK for Go. The following guidance for creating samples helps ensure consistency across our many supported services.

* Every service in `azure-sdk-for-go/services/...` is eligible for a top-level folder in `azure-sdk-for-go-samples` (this repo).
* Each top-level folder in `azure-sdk-for-go-samples` is a library of common operations typically executed for that service. Operations are associated with entities in the service's API and included in a file named for that entity. For example, the `storage/` directory is arranged as described below.
* Each top-level folder should contain at least one testable file named `<service>_test.go` which exercises all methods in the directory.
* An example top-level directory for Azure Storage follows. The leaf nodes are method names within a file.

```
  storage/
      account.go
          CreateAccount
          DeleteAccount
          GetAccount
          UpdateAccount
      block_blob.go
          CreateBlockBlob
          CreateBlockBlobWithStream
          DeleteBlockBlob
      container.go
          ...
      file.go
          ...
      page_blob.go
          ...
      queue.go
      storage_test.go
          Example
          ExampleFile
```

* Samples for one service should utilize methods from other samples for non-essential operations. For example, compute should utilize operations from network to deploy a network for a VM; and all samples should utilize operations from `iam/` for authentication.
* All samples should use the same conventions for naming and using environment variables. This convention currently is:
  * All env vars used across the samples repo should be listed in the root `.env.tpl`. This allows a user to set all env vars in one place and run `go test ./...`.
  * All top-level directories should also include a `.env.tpl` file with only the vars needed for that particular sample. This allows running just one set of tests via `go test ./sql/` and the like.
  * Env vars should be named `AZURE_<SERVICE>_<VARNAME>`, e.g. `AZURE_STORAGE_ACCOUNTNAME` and `AZURE_VNET_NAME`.