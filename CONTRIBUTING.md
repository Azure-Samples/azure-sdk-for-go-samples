# Contents

* [Guidance for writing new samples](#guidance)
* [Contributor License Agreement (CLA)](#cla)
* [Code of Conduct](#code-of-conduct)
* [Submission Guidelines](#submit)

# <a name="guidance"></a> Guidance for writing new samples

This repo provides samples to help developers understand how to interact with Azure services using the Azure SDK for Go. The following guidance for creating samples helps ensure consistency across our many supported services.

* Any service in [`azure-sdk-for-go/services/...`][1] is eligible for a top-level folder in this repo (`azure-sdk-for-go-samples`).
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
  * Env vars should be named `AZ_<SERVICE>_<VARNAME>`, e.g. `AZ_STORAGE_ACCOUNTNAME` and `AZ_VNET_NAME`.

[1]: https://github.com/Azure/azure-sdk-for-go/tree/master/services


## <a name="cla"></a> Contributor License Agreement (CLA)

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

## <a name="code-of-conduct"></a> Code of Conduct

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

 - [Code of Conduct](#coc)
 - [Issues and Bugs](#issue)
 - [Feature Requests](#feature)
 - [Submission Guidelines](#submit)

## <a name="issue"></a> Found an Issue?

If you find a bug in the source code or a mistake in the documentation, you can help us by
[submitting an issue](#submit-issue) to the GitHub Repository. Even better, you can
[submit a Pull Request](#submit-pr) with a fix.

## <a name="feature"></a> Want a Feature?

You can *request* a new feature by [submitting an issue](#submit-issue) to the GitHub
Repository. If you would like to *implement* a new feature, please submit an issue with
a proposal for your work first, to be sure that we can use it.

* **Small Features** can be crafted and directly [submitted as a Pull Request](#submit-pr).

## <a name="submit"></a> Submission Guidelines

### <a name="submit-issue"></a> Submitting an Issue
Before you submit an issue, search the archive, maybe your question was already answered.

If your issue appears to be a bug, and hasn't been reported, open a new issue.
Help us to maximize the effort we can spend fixing issues and adding new
features, by not reporting duplicate issues.  Providing the following information will increase the
chances of your issue being dealt with quickly:

* **Overview of the Issue** - if an error is being thrown a non-minified stack trace helps
* **Version** - what version is affected (e.g. 0.1.2)
* **Motivation for or Use Case** - explain what are you trying to do and why the current behavior is a bug for you
* **Browsers and Operating System** - is this a problem with all browsers?
* **Reproduce the Error** - provide a live example or a unambiguous set of steps
* **Related Issues** - has a similar issue been reported before?
* **Suggest a Fix** - if you can't fix the bug yourself, perhaps you can point to what might be
  causing the problem (line of code or commit)
You can file new issues by providing the above information at the corresponding repository's issues link: https://github.com/Azure-Samples/azure-sdk-for-go-samples/issues/new].

### <a name="submit-pr"></a> Submitting a Pull Request (PR)

Before you submit your Pull Request (PR) consider the following guidelines:

* Search [this repository](https://github.com/Azure-Samples/azure-sdk-for-go-samples/pulls) for an open or closed PR
  that relates to your submission. You don't want to duplicate effort.
* [Fork](https://github.com/Azure-Samples/azure-sdk-for-go-samples/fork) this repo.
* Write and commit your changes using a descriptive commit message.
* Push your changes to your fork.
* Create a Pull Request from the GitHub web interface.
* If changes are needed:

    * Make required updates in your local fork.
    * Rebase your fork against the main repo and force-push the new series of commits to your GitHub repo:

    ```shell
    $ git rebase origin/master
    $ git push --force
    ```

That's it! Thank you for your contribution!
