# ARM Balancer

A client-side connection manager for Azure Resource Manager.

## Why?

ARM request throttling is scoped to the specific instance of ARM that a connection lands on.
This serves to reduce the risk of a noisy client impacting the performance of other requests handled concurrently by that particular instance without requiring coordination between instances.

HTTP1.1 clients commonly use pooled TCP connections to provide concurrency.
But HTTP2 allows a single connection to handle many concurrent requests.
Conforming client implementations will only open a second connection when the concurrency limit advertised by the server would be exceeded.

This poses a problem for ARM consumers using HTTP2: requests that were previously distributed across several ARM instances will now be sent to only one.

## Design

- Multiple connections are established with ARM, forming a simple client-side load balancer
- Connections are re-established when they receive a "ratelimit-remaining" header below a certain threshold

This scheme avoids throttling by proactively redistributing load across ARM instances.
Performance under high concurrency may also improve relative to HTTP1.1 since the pool of connections can easily be made larger than common HTTP client defaults.

## Usage

```go
armresources.NewClient("{{subscriptionID}}", cred, &arm.ClientOptions{
	ClientOptions: policy.ClientOptions{
		Transport: &http.Client{
			Transport: armbalancer.New(armbalancer.Options{}),
		},
	},
})
```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
