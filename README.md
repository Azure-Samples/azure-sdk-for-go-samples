Simple samples for using Azure services with Go.

## To run

1. Copy [`.env.tpl`](./.env.tpl): `cp .env.tpl .env`, and fill in appropriately.
   Subscription, tenant, and client IDs and secrets aren't needed for Device Flow auth.
1. Optional: run `dep ensure`.
1. Run: `go test ./storage/`, `go test ./network/`, or if you're feeling lucky `go test ./...`

To use Device Flow rather than Service Principal, replace `iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)` with `iam.GetResourceManagementToken(iam.OAuthGrantTypeDeviceFlow)` throughout codebase. Currently, Device Flow relies on an app ID from joshgav's tenant.

Keep created resources by setting env var: `AZURE_KEEP_SAMPLE_RESOURCES=1`.
