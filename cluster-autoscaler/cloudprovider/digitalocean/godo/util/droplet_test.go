package util

import (
	"context"

	"golang.org/x/oauth2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/digitalocean/godo"
)

func ExampleWaitForActive() {
	// build client
	pat := "mytoken"
	token := &oauth2.Token{AccessToken: pat}
	t := oauth2.StaticTokenSource(token)

	ctx := context.TODO()
	oauthClient := oauth2.NewClient(ctx, t)
	client := godo.NewClient(oauthClient)

	// create your droplet and retrieve the create action uri
	uri := "https://api.digitalocean.com/v2/actions/xxxxxxxx"

	// block until until the action is complete
	err := WaitForActive(ctx, client, uri)
	if err != nil {
		panic(err)
	}
}
