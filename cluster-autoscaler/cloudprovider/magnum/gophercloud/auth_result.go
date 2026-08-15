package gophercloud

/*
AuthResult is the result from the request that was used to obtain a provider
client's Keystone token. It is returned from ProviderClient.GetAuthResult().

The following types satisfy this interface:

	k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v2/tokens.CreateResult
	k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v3/tokens.CreateResult

Usage example:

	import (
		"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
		tokens2 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v2/tokens"
		tokens3 "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/identity/v3/tokens"
	)

	func GetAuthenticatedUserID(providerClient *gophercloud.ProviderClient) (string, error) {
		r := providerClient.GetAuthResult()
		if r == nil {
			//ProviderClient did not use openstack.Authenticate(), e.g. because token
			//was set manually with ProviderClient.SetToken()
			return "", errors.New("no AuthResult available")
		}
		switch r := r.(type) {
		case tokens2.CreateResult:
			u, err := r.ExtractUser()
			if err != nil {
				return "", err
			}
			return u.ID, nil
		case tokens3.CreateResult:
			u, err := r.ExtractUser()
			if err != nil {
				return "", err
			}
			return u.ID, nil
		default:
			panic(fmt.Sprintf("got unexpected AuthResult type %t", r))
		}
	}

Both implementing types share a lot of methods by name, like ExtractUser() in
this example. But those methods cannot be part of the AuthResult interface
because the return types are different (in this case, type tokens2.User vs.
type tokens3.User).
*/
type AuthResult interface {
	ExtractTokenID() (string, error)
}
