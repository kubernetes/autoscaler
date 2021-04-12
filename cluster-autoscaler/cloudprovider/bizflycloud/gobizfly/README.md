# bizfly-client-go

# Example

```go
package main

import (
	"context"
	"log"
	"time"

	gobizfly "github.com/bizflycloud/bizfly-client-go"
)

const (
	host     = "https://manage.bizflycloud.vn"
	username = "cuonglm@vccloud.vn"
	password = "foobar"
)

func main() {
	client, err := gobizfly.NewClient(
		gobizfly.WithAPIUrl(host),
		gobizfly.WithTenantName(username),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	tok, err := client.Token.Create(ctx, &gobizfly.TokenCreateRequest{Username: username, Password: password})
	if err != nil {
		log.Fatal(err)
	}
	client.SetKeystoneToken(tok.KeystoneToken)

	lbs, err := client.LoadBalancer.List(ctx, &gobizfly.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v\n", lbs)
}
```

For other usages, see test files.
