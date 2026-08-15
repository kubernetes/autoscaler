package gobrightbox

// ConfigMap represents a config map
// https://api.gb1.brightbox.com/1.0/#config_maps
type ConfigMap struct {
	Id   string                 `json:"id"`
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

// ConfigMapOptions is used in combination with CreateConfigMap and
// UpdateConfigMap to create and update config maps
type ConfigMapOptions struct {
	Id   string                  `json:"-"`
	Name *string                 `json:"name,omitempty"`
	Data *map[string]interface{} `json:"data,omitempty"`
}

// ConfigMaps retrieves a list of all config maps
func (c *Client) ConfigMaps() ([]ConfigMap, error) {
	var configMaps []ConfigMap
	_, err := c.MakeApiRequest("GET", "/1.0/config_maps", nil, &configMaps)
	if err != nil {
		return nil, err
	}
	return configMaps, err
}

// ConfigMap retrieves a detailed view on one config map
func (c *Client) ConfigMap(identifier string) (*ConfigMap, error) {
	configMap := new(ConfigMap)
	_, err := c.MakeApiRequest("GET", "/1.0/config_maps/"+identifier, nil, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, err
}

// CreateConfigMap creates a new config map
//
// It takes an instance of ConfigMapOptions. Not all attributes can be
// specified at create time (such as Id, which is allocated for you).
func (c *Client) CreateConfigMap(newConfigMap *ConfigMapOptions) (*ConfigMap, error) {
	configMap := new(ConfigMap)
	_, err := c.MakeApiRequest("POST", "/1.0/config_maps", newConfigMap, &configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// UpdateConfigMap updates an existing config maps's attributes. Not all
// attributes can be changed (such as Id).
//
// Specify the config map you want to update using the ConfigMapOptions Id
// field.
func (c *Client) UpdateConfigMap(updateConfigMap *ConfigMapOptions) (*ConfigMap, error) {
	configMap := new(ConfigMap)
	_, err := c.MakeApiRequest("PUT", "/1.0/config_maps/"+updateConfigMap.Id, updateConfigMap, &configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// DestroyConfigMap destroys an existing config map
func (c *Client) DestroyConfigMap(identifier string) error {
	_, err := c.MakeApiRequest("DELETE", "/1.0/config_maps/"+identifier, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
