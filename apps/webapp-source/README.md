# Web App from source

**NOTE**: support for Go web apps from source is in preview and not supported
for production apps yet. The feature is hidden behind a flag which the scripts
here use.

Continuously build and deploy a Go web app using Azure App Service and GitHub,
no container required.

This package includes a basic Go web app and scripts to set up infrastructure
for it in Azure.

Requires [Azure CLI][].

## Try It!

1. Copy the contents of this package to a new repo.
1. Set configuration in a local `.env` file in the package root by copying
   `.env.tpl` to `.env`.
1. Change `AZURE_BASE_NAME` to something relatively unique to you; for example
   you might include your name.
1. Change `REPO_NAME` to your org and repo in GitHub, e.g. `joshgav/go-sample`.
1. Run ./[setup.sh][] to set up an App Service web app connected to the
   specified repo. Don't forget to `git push` your code there too!

**NOTE**: GitHub sync requires a [GitHub personal access
token](https://github.com/settings/tokens); you need to get one from the linked
page and set it in a local environment variable `GH_TOKEN`, e.g. `export
GH_TOKEN=mylongtokenstring`. You can also add it to your local `.env` file for
persistence.

To test continuous integration, now make a change and `git push` it to your
repo. The GitHub sync task should detect the change and rebuild and refresh
your web app.

## Details

[setup.sh][] ensures an Azure resource group, app service plan, and
source-based web app are provisioned and connected in the subscription
currently logged in to [Azure CLI][].

It uses the following environment variables to choose names:

* REPO\_NAME: GitHub repo in form `organization/repo`.
* AZURE\_BASE\_NAME: Prefix for Azure resources.
* AZURE\_DEFAULT\_LOCATION: Location for Azure resources. 
* GH\_TOKEN: A GitHub [personal access token](https://github.com/settings/tokens)

These names can be specified in a .env file in the root of the package. If a
`.env` file isn't found, `.env.tpl` is copied to `.env` and used.

Explicit parameters can also be passed, see comments at beginning of
[setup.sh][] for details.

[Azure CLI]: https://github.com/Azure/azure-cli
[setup.sh]: ./setup.sh
