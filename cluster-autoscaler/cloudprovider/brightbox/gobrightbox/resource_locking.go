package gobrightbox

import (
	"fmt"
)

func resourcePath(resource interface{}) (string, error) {
	switch resource := resource.(type) {
	default:
		return "", fmt.Errorf("Unknown resource type %s", resource)
	case *Server:
		return "servers/" + resource.Id, nil
	case Server:
		return "servers/" + resource.Id, nil
	case *Image:
		return "images/" + resource.Id, nil
	case Image:
		return "images/" + resource.Id, nil
	case *LoadBalancer:
		return "load_balancers/" + resource.Id, nil
	case LoadBalancer:
		return "load_balancers/" + resource.Id, nil
	case *DatabaseServer:
		return "database_servers/" + resource.Id, nil
	case DatabaseServer:
		return "database_servers/" + resource.Id, nil
	case *ApiClient:
		return "api_clients/" + resource.Id, nil
	case ApiClient:
		return "api_clients/" + resource.Id, nil
	}
}

// LockResource locks a resource against destroy requests. Support brightbox.Server, brightbox.Image, brightbox.DatabaseServer and brightbox.LoadBalancer
func (c *Client) LockResource(resource interface{}) error {
	rpath, err := resourcePath(resource)
	if err != nil {
		return err
	}
	_, err = c.MakeApiRequest("PUT", fmt.Sprintf("/1.0/%s/lock_resource", rpath), nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// UnLockResource unlocks a resource, renabling destroy requests
func (c *Client) UnLockResource(resource interface{}) error {
	rpath, err := resourcePath(resource)
	if err != nil {
		return err
	}
	_, err = c.MakeApiRequest("PUT", fmt.Sprintf("/1.0/%s/unlock_resource", rpath), nil, nil)
	if err != nil {
		return err
	}
	return nil
}
