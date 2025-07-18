## Performance Utility

Downloads a test file from a S3 bucket using the SDK's S3 download manager. Allows passing
in a custom configuration for the HTTP client and SDK's Download Manager behavior.

## Build
```sh
go test -tags "integration perftest" -c -o download.test ./manager/internal/integration/download
```

## Usage Example:
```sh
AWS_REGION=us-west-2 AWS_PROFILE=aws-go-sdk-team-test ./download.test \
-test.bench=. \
-test.benchmem \
-test.benchtime 1x \
-bucket aws-sdk-go-data \
-client.idle-conns 1000 \
-client.idle-conns-host 300 \
-client.timeout.connect=1s \
-client.timeout.response-header=1s
```
