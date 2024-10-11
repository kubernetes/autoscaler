# Oracle Cloud Infrastructure Golang SDK
[![wercker status](https://app.wercker.com/status/09bc4818e7b1d70b04285331a9bdbc41/s/master "wercker status")](https://app.wercker.com/project/byKey/09bc4818e7b1d70b04285331a9bdbc41)

This is the Go SDK for Oracle Cloud Infrastructure. This project is open source and maintained by Oracle Corp. 
The home page for the project is [here](https://godoc.org/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/).

## Survey
Are you a Developer using the OCI SDK? If so, please fill out our survey to help us make the OCI SDK better for you. Click [here](https://oracle.questionpro.com/t/APeMlZka26?custom3=pkg) for the survey page.


## Dependencies
- Install [Go programming language](https://golang.org/dl/), Go1.17, 1.18, 1.19, 1.20, and 1.21 are supported By OCI Go SDK.
- Install [GNU Make](https://www.gnu.org/software/make/), using the package manager or binary distribution tool appropriate for your platform.
 
## Versioning
- The breaking changes in service client APIs will no longer result in a major version bump (x+1.y.z relative to the last published version). Instead, we will bump the minor version of the SDK (x.y+1.z relative to the last published version).
- If there are no breaking changes in a release, we will bump the patch version of the SDK (x.y.z+1 relative to the last published version).
We will continue to announce any breaking changes in a new version under the Breaking Changes section in the release changelog and on the Github release page [https://github.com/oracle/oci-go-sdk/releases].
- However, breaking changes in the SDK's common module will continue to result in a major version bump (x+1.y.z relative to the last published version). That said, we'll continue to maintain backward compatibility as much as possible to minimize the effort involved in upgrading the SDK version used by your code.

## Installing
If you want to install the SDK under $GOPATH, you can use `go get` to retrieve the SDK:
```
go get -u github.com/oracle/oci-go-sdk
```
If you are using Go modules, you can install by running the following command within a folder containing a `go.mod` file:
```
go get -d github.com/oracle/oci-go-sdk/v49@latest
```
The latest major version (for example `v49`) can be identified on the [Github releases page](https://github.com/oracle/oci-go-sdk/releases).

Alternatively, you can install a specific version (supported from `v25.0.0` on):
```
go get -d github.com/oracle/oci-go-sdk/v49@v49.1.0
```
Run `go mod tidy`

In your project, you also need to ensure the import paths contain the correct major-version:
```
import "github.com/oracle/oci-go-sdk/v49/common"  // or whatever major version you're using
```

## Working with the Go SDK
To start working with the Go SDK, you import the service package, create a client, and then use that client to make calls.

### Configuring 
Before using the SDK, set up a config file with the required credentials. See [SDK and Tool Configuration](https://docs.cloud.oracle.com/Content/API/Concepts/sdkconfig.htm) for instructions.

Note that the Go SDK does not support profile inheritance or defining custom values in the configuration file.

Once a config file has been setup, call `common.DefaultConfigProvider()` function as follows:

 ```go
 // Import necessary packages
 import (
	"github.com/oracle/oci-go-sdk/v49/common"
	"github.com/oracle/oci-go-sdk/v49/identity" // Identity or any other service you wish to make requests to
)
 
 //...
 
configProvider := common.DefaultConfigProvider()
```

 Or, to configure the SDK programmatically instead, implement the `ConfigurationProvider` interface shown below:
 ```go
// ConfigurationProvider wraps information about the account owner
type ConfigurationProvider interface {
	KeyProvider
	TenancyOCID() (string, error)
	UserOCID() (string, error)
	KeyFingerprint() (string, error)
	Region() (string, error)
	// AuthType() is used for specify the needed auth type, like UserPrincipal, InstancePrincipal, etc. AuthConfig is used for getting auth related paras in config file.
	AuthType() (AuthConfig, error)
}
```
Or simply use one of  structs exposed by the `oci-go-sdk` that already implement the above [interface](https://godoc.org/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/common#ConfigurationProvider)

### Making a Request
To make a request to an Oracle Cloud Infrastructure service, create a client for the service and then use the client to call a function from the service.

- *Creating a client*: All packages provide a function to create clients, using the naming convention `New<ServiceName>ClientWithConfigurationProvider`,
such as `NewVirtualNetworkClientWithConfigurationProvider` or `NewIdentityClientWithConfigurationProvider`. To create a new client, 
pass a struct that conforms to the `ConfigurationProvider` interface, or use the `DefaultConfigProvider()` function in the common package.

For example: 
```go
config := common.DefaultConfigProvider()
client, err := identity.NewIdentityClientWithConfigurationProvider(config)
if err != nil { 
     panic(err)
}
```

- *Making calls*: After successfully creating a client, requests can now be made to the service. Generally all functions associated with an operation
accept [`context.Context`](https://golang.org/pkg/context/) and a struct that wraps all input parameters. The functions then return a response struct
that contains the desired data, and an error struct that describes the error if an error occurs.

For example:
```go
id := "your_group_id"
response, err := client.GetGroup(context.Background(), identity.GetGroupRequest{GroupId:&id})
if err != nil {
	//Something happened
	panic(err)
}
//Process the data in response struct
fmt.Println("Group's name is:", response.Name)
```

- *Expect header*: Some services specified "PUT/POST" requests add Expect 100-continue header by default, if it is not expected, please explicitly set the env var:
```sh
export OCI_GOSDK_USING_EXPECT_HEADER=FALSE
```

- *Circuit Breaker*: By default, circuit breaker feature is enabled, if it is not expected, please explicitly set the env var:
```sh
export OCI_SDK_DEFAULT_CIRCUITBREAKER_ENABLED=FALSE
```
- *Cicuit Breaker*: Circuit Breaker error message includes a set of previous failed responses. By default, the number of the failed responses is set to 5. It can be explicitly set using the env var:
```sh
export OCI_SDK_CIRCUITBREAKER_NUM_HISTORY_RESPONSE=<int value>
``` 

## Organization of the SDK
The `oci-go-sdk` contains the following:
- **Service packages**: All packages except `common` and any other package found inside `cmd`. These packages represent 
the Oracle Cloud Infrastructure services supported by the Go SDK. Each package represents a service. 
These packages include methods to interact with the service, structs that model 
input and output parameters, and a client struct that acts as receiver for the above methods.

- **Common package**: Found in the `common` directory. The common package provides supporting functions and structs used by service packages.
Includes HTTP request/response (de)serialization, request signing, JSON parsing, pointer to reference and other helper functions. Most of the functions
in this package are meant to be used by the service packages.

- **cmd**: Internal tools used by the `oci-go-sdk`.

## Examples
Examples can be found [here](https://github.com/oracle/oci-go-sdk/tree/master/example)

## Documentation
Full documentation can be found [on the godocs site](https://godoc.org/k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/vendor-internal/github.com/oracle/oci-go-sdk/v65/).

## Help
* The [Issues](https://github.com/oracle/oci-go-sdk/issues) page of this GitHub repository.
* [Stack Overflow](https://stackoverflow.com/), use the [oracle-cloud-infrastructure](https://stackoverflow.com/questions/tagged/oracle-cloud-infrastructure) and [oci-go-sdk](https://stackoverflow.com/questions/tagged/oci-go-sdk) tags in your post.
* [Developer Tools](https://community.oracle.com/community/cloud_computing/bare-metal/content?filterID=contentstatus%5Bpublished%5D~category%5Bdeveloper-tools%5D&filterID=contentstatus%5Bpublished%5D~objecttype~objecttype%5Bthread%5D) of the Oracle Cloud forums.
* [My Oracle Support](https://support.oracle.com).


## Contributing
This project welcomes contributions from the community. Before submitting a pull request, please [review our contribution guide](./CONTRIBUTING.md)

## Security
Please consult the [security guide](./SECURITY.md) for our responsible security vulnerability disclosure process

## License
Copyright (c) 2016, 2023, Oracle and/or its affiliates.  All rights reserved.
This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl
or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.

See [LICENSE](/LICENSE.txt) for more details.

## Changes
See [CHANGELOG](/CHANGELOG.md).

## Known Issues
You can find information on any known issues with the SDK here and under the [Issues](https://github.com/oracle/oci-go-sdk/issues) tab of this project's GitHub repository.

## Building and Testing
### Dev Dependencies
For Go versions below 1.17

- Install [Testify](https://github.com/stretchr/testify) with the command:
```sh
go get github.com/stretchr/testify
```
- Install [gobreaker](https://github.com/sony/gobreaker) with the command:
```sh
go get github.com/sony/gobreaker
```
- Install [flock](https://github.com/gofrs/flock) with the command:
```sh
go get github.com/gofrs/flock
```
- Install [go lint](https://github.com/golang/lint) with the command:
```
go get -u golang.org/x/lint/golint
```

For Go versions 1.17 and above

- Install [Testify](https://github.com/stretchr/testify) with the command:
```sh
go install github.com/stretchr/testify
```
- Install [gobreaker](https://github.com/sony/gobreaker) with the command:
```sh
go install github.com/sony/gobreaker
```
- Install [flock](https://github.com/gofrs/flock) with the command:
```sh
go install github.com/gofrs/flock
```
- Install [go lint](https://github.com/golang/lint) with the command:
```
go install github.com/golang/lint/golint
```
- Install [staticcheck](https://github.com/dominikh/go-tools) with the command:
```
go install honnef.co/go/tools/cmd/staticcheck@2023.1.7
```

### Linting and Staticcheck
Linting (performed by golint) can be done with the following command:

```
make lint
```
Linting will perform a number of formatting changes across the code base.


```
make static-check
```
This command is also run by the make build and make test commands.
Staticcheck will provide formatting warnings but will not make any changes to any files.
You can also cause staticcheck to be run before calls to "git commit" with the pre-commit plugin.

```
pre-commit install
```
You can install pre-commit itself, you can use your package manager of choice, such as

```
brew install pre-commit
```

### Build
Building is provided by the make file at the root of the project. To build the project execute.

```
make build
```

To run the tests:
```
make test
```
