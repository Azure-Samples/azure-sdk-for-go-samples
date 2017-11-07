Simple samples for using az services with go

## To run

### Option 1: Service Principal

1. Copy the .env template: `cp .env.template .env`
1. Fill in your own .env appropriately.
1. Run: `glide install` (to use the `dev` branch of the SDK repo)
1. Run: `go run main.go`

If you want to keep the created resources, run `AZURE_KEEP_SAMPLE_RESOURCES=1 go run main.go`.

### Option 2: Device Flow

Currently uses an app ID from joshgav's tenant.

TODO(joshgav): add command-line flags to choose device flow

1. You'll still need to specify a resource group name, database name, etc. in the .env file, so copy the .env template: `cp .env.template .env` and fill it in.
1. `glide install`
1. Replace `common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)` with `common.GetResourceManagementToken(common.OAuthGrantTypeDeviceFlow)` throughout the codebase. 

