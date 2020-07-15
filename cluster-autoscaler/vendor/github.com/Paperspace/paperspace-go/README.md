# go-paperspace

## Usage
```go
package main

import (
    paperspace "github.com/Paperspace/paperspace-go"
)

func getClient() *paperspace.Client {
    apiBackend := paperspace.NewAPIBackend()
    if p.BaseURL != "" {
        apiBackend.BaseURL = p.BaseURL
    }
    if os.Getenv("PAPERSPACE_BASEURL") != "" {
        apiBackend.BaseURL = os.Getenv("PAPERSPACE_BASEURL")
    }
    apiBackend.Debug = p.Debug
    if os.Getenv("PAPERSPACE_DEBUG") != "" {
        apiBackend.Debug = true
    }
    apiBackend.DebugBody = p.DebugBody
    if os.Getenv("PAPERSPACE_DEBUG_BODY") != "" {
        apiBackend.DebugBody = true
    }
    client := paperspace.NewClientWithBackend(paperspace.Backend(apiBackend))
    client.APIKey = p.APIKey
    return client
}
```
