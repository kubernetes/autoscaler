# go-paperspace

## Usage
```go
package main

import (
    paperspace "github.com/Paperspace/paperspace-go"
)

func getClient() *paperspace.Client {
    client := paperspace.NewClient()
    client.APIKey = p.APIKey
    return client
}
```

## Environment Variables
- PAPERSPACE_APIKEY: Paperspace API key
- PAPERSPACE_BASEURL: Paperspace API url
- PAPERSPACE_DEBUG: Enable debugging
- PAPERSPACE_DEBUG_BODY: Enable debug for response body