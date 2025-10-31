/*
Package customizations provides customizations for the Amazon S3-Control API client.

This package provides support for following S3-Control customizations

	BackfillInput Middleware: validates and backfills data from an ARN resource into a copy of operation input.

	ProcessOutpostID Middleware: applied on CreateBucket, ListRegionalBuckets operation, triggers a custom endpoint generation flow.

	ProcessARN Middleware: processes an ARN if provided as input and updates the endpoint as per the arn type.

	UpdateEndpoint Middleware: resolves a custom endpoint as per s3-control config options.

# Dualstack support

By default dualstack support for s3-control client is disabled. By enabling `UseDualstack`
option on s3-control client, you can enable dualstack endpoint support.

# Endpoint customizations

Customizations to lookup ARN, backfill input, process outpost id, process ARN
needs to happen before request serialization. UpdateEndpoint middleware which mutates
resources based on Options such as UseDualstack for modifying resolved endpoint
are executed after request serialization.

	Middleware layering:

	Initialize : HTTP Request -> ARN Lookup -> BackfillInput -> Input-Validation -> Serialize step

	Serialize : HTTP Request -> Process-OutpostID -> Process ARN -> operation serializer -> Update-Endpoint customization -> next middleware

Customization option:

	UseARNRegion (Disabled by Default)

	UseDualstack (Disabled by Default)
*/
package customizations
