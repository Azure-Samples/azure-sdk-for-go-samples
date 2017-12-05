Simple samples for using Azure services with Go.

## To run

1. Run `dep ensure`

### Option 1: Service Principal

1. Copy the .env template: `cp .env.template .env`
1. Fill in your own .env appropriately.
1. Run: `go run examples/mssql/main.go` or `go run examples/network/main.go`, etc.
  
### Option 2: Device Flow

Currently uses an app ID from joshgav's tenant.

1. Replace `common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)` with `common.GetResourceManagementToken(common.OAuthGrantTypeDeviceFlow)` throughout the codebase. TODO(joshgav): make this configurable at runtime (e.g. a flag).
1. Copy the .env template: `cp .env.template .env`.
1. Fill in names in .env. Subscription, tenant, and client IDs and secrets aren't needed.
1. Run: `go run examples/mssql/main.go` or `go run examples/network/main.go`, etc.

## Notes
  
Keep created resources by passing flag: `--keepResources`.

