package skewer

import (
	"context"
	"fmt"
	"strings"
)

// Config contains configuration options for a cache.
type Config struct {
	location string
	filter   string
	client   client
}

// Cache stores a list of known skus, possibly fetched with a provided client
type Cache struct {
	config *Config
	data   []SKU
}

// Option describes functional options to customize the listing behavior of the cache.
type Option func(c *Config) (*Config, error)

// WithLocation is a functional option to filter skus by location
func WithLocation(location string) Option {
	return func(c *Config) (*Config, error) {
		c.location = location
		c.filter = fmt.Sprintf("location eq '%s'", location)
		return c, nil
	}
}

// ErrClientNil will be returned when a user attempts to create a cache
// without a client and use it.
type ErrClientNil struct {
}

func (e *ErrClientNil) Error() string {
	return "cache requires a client provided by functional options to refresh"
}

// ErrClientNotNil will be returned when a user attempts to set two
// clients on the same cache.
type ErrClientNotNil struct {
}

func (e *ErrClientNotNil) Error() string {
	return "only provide one client option when instantiating a cache"
}

// WithClient is a functional option to use a cache
// backed by a client meeting the skewer signature.
func WithClient(client client) Option {
	return func(c *Config) (*Config, error) {
		if c.client != nil {
			return nil, &ErrClientNotNil{}
		}
		c.client = client
		return c, nil
	}
}

// WithResourceClient is a functional option to use a cache
// backed by a ResourceClient.
func WithResourceClient(client ResourceClient) Option {
	return func(c *Config) (*Config, error) {
		if c.client != nil {
			return nil, &ErrClientNotNil{}
		}
		c.client = newWrappedResourceClient(client)
		return c, nil
	}
}

// WithResourceProviderClient is a functional option to use a cache
// backed by a ResourceProviderClient.
func WithResourceProviderClient(client ResourceProviderClient) Option {
	return func(c *Config) (*Config, error) {
		if c.client != nil {
			return nil, &ErrClientNotNil{}
		}
		resourceClient := newWrappedResourceProviderClient(client)
		c.client = newWrappedResourceClient(resourceClient)
		return c, nil
	}
}

// NewCacheFunc describes the live cache instantiation signature. Used
// for testing.
type NewCacheFunc func(ctx context.Context, opts ...Option) (*Cache, error)

// NewCache instantiates a cache of resource sku data with a ResourceClient
// client, optionally with additional filtering by location. The
// accepted client interface matches the real Azure clients (it returns
// a paginated iterator).
func NewCache(ctx context.Context, opts ...Option) (*Cache, error) {
	config := &Config{}

	for _, optionFn := range opts {
		var err error
		if config, err = optionFn(config); err != nil {
			return nil, err
		}
	}

	if config.client == nil {
		return nil, &ErrClientNil{}
	}

	c := &Cache{
		config: config,
	}

	if err := c.refresh(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

// NewStaticCache initializes a cache with data and no ability to refresh. Used for testing.
func NewStaticCache(data []SKU, opts ...Option) (*Cache, error) {
	config := &Config{}

	for _, optionFn := range opts {
		var err error
		if config, err = optionFn(config); err != nil {
			return nil, err
		}
	}

	c := &Cache{
		data:   data,
		config: config,
	}

	return c, nil
}

func (c *Cache) refresh(ctx context.Context) error {
	data, err := c.config.client.List(ctx, c.config.filter)
	if err != nil {
		return err
	}

	c.data = Wrap(data)

	return nil
}

// ErrMultipleSKUsMatch will be returned when multiple skus match a
// fully qualified triple of resource type, location and name. This should usually not happen.
type ErrMultipleSKUsMatch struct {
	Name     string
	Location string
	Type     string
}

func (e *ErrMultipleSKUsMatch) Error() string {
	return fmt.Sprintf("found multiple skus matching type: %s, name %s, and location %s", e.Type, e.Name, e.Location)
}

// ErrSKUNotFound will be returned when no skus match a fully qualified
// triple of resource type, location and name. The SKU may not exist.
type ErrSKUNotFound struct {
	Name     string
	Location string
	Type     string
}

func (e *ErrSKUNotFound) Error() string {
	return fmt.Sprintf("failed to find any skus matching type: %s, name %s, and location %s", e.Type, e.Name, e.Location)
}

// Get returns the first matching resource of a given name and type in a location.
func (c *Cache) Get(ctx context.Context, name, resourceType, location string) (SKU, error) {
	filtered := Filter(c.data, []FilterFn{
		ResourceTypeFilter(resourceType),
		NameFilter(name),
		LocationFilter(location),
	}...)

	if len(filtered) > 1 {
		return SKU{}, &ErrMultipleSKUsMatch{
			Name:     name,
			Location: location,
			Type:     resourceType,
		}
	}

	if len(filtered) < 1 {
		return SKU{}, &ErrSKUNotFound{
			Name:     name,
			Location: location,
			Type:     resourceType,
		}
	}

	return filtered[0], nil
}

// List returns all resource types for this location.
func (c *Cache) List(ctx context.Context, filters ...FilterFn) []SKU {
	return Filter(c.data, filters...)
}

// GetVirtualMachines returns the list of all virtual machines *SKUs in a given azure location.
func (c *Cache) GetVirtualMachines(ctx context.Context) []SKU {
	return Filter(c.data, ResourceTypeFilter(VirtualMachines))
}

// GetVirtualMachineAvailabilityZones returns all virtual machine zones available in a given location.
func (c *Cache) GetVirtualMachineAvailabilityZones(ctx context.Context) []string {
	return c.GetAvailabilityZones(ctx, ResourceTypeFilter(VirtualMachines))
}

// GetVirtualMachineAvailabilityZonesForSize returns all virtual machine zones available in a given location.
func (c *Cache) GetVirtualMachineAvailabilityZonesForSize(ctx context.Context, size string) []string {
	return c.GetAvailabilityZones(ctx, ResourceTypeFilter(VirtualMachines), NameFilter(size))
}

// GetAvailabilityZones returns the list of all availability zones in a given azure location.
func (c *Cache) GetAvailabilityZones(ctx context.Context, filters ...FilterFn) []string {
	allZones := make(map[string]bool)

	Map(c.data, func(s *SKU) SKU {
		if All(s, filters) {
			for zone := range s.AvailabilityZones(c.config.location) {
				allZones[zone] = true
			}
		}
		return SKU{}
	})

	result := make([]string, 0, len(allZones))
	for zone := range allZones {
		result = append(result, zone)
	}

	return result
}

// Equal compares two configs.
func (c *Config) Equal(other *Config) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil && other != nil {
		return false
	}
	if c != nil && other == nil {
		return false
	}
	return c.location == other.location &&
		c.filter == other.filter
}

// Equal compares two caches.
func (c *Cache) Equal(other *Cache) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil && other != nil {
		return false
	}
	if c != nil && other == nil {
		return false
	}
	if c != nil && other != nil {
		return c.config.Equal(other.config)
	}
	if len(c.data) != len(other.data) {
		return false
	}
	for i := range c.data {
		if c.data[i] != other.data[i] {
			return false
		}
	}
	return true
}

// All returns true if the provided sku meets all provided conditions.
func All(sku *SKU, conditions []FilterFn) bool {
	for _, condition := range conditions {
		if !condition(sku) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing all values in the slice that
// satisfy all filterFn predicates.
func Filter(skus []SKU, filterFn ...FilterFn) []SKU {
	if skus == nil {
		return nil
	}

	filtered := make([]SKU, 0)
	for i := range skus {
		if All(&skus[i], filterFn) {
			filtered = append(filtered, skus[i])
		}
	}

	return filtered
}

// Map returns a new slice containing the results of applying the
// mapFn to each value in the original slice.
func Map(skus []SKU, fn MapFn) []SKU {
	if skus == nil {
		return nil
	}

	mapped := make([]SKU, 0, len(skus))
	for i := range skus {
		mapped = append(mapped, fn(&skus[i]))
	}

	return mapped
}

// FilterFn is a convenience type for filtering.
type FilterFn func(*SKU) bool

// ResourceTypeFilter produces a filter function for any resource type.
func ResourceTypeFilter(resourceType string) func(*SKU) bool {
	return func(s *SKU) bool {
		return s.IsResourceType(resourceType)
	}
}

// NameFilter produces a filter function for the name of a resource sku.
func NameFilter(name string) func(*SKU) bool {
	return func(s *SKU) bool {
		return strings.EqualFold(s.GetName(), name)
	}
}

// LocationFilter matches against a SKU listing the given location
func LocationFilter(location string) func(*SKU) bool {
	return func(s *SKU) bool {
		return s.HasLocation(normalizeLocation(location))
	}
}

// UnsafeLocationFilter produces a filter function for the location of a
// resource sku.
// This function dangerously ignores all SKUS without a properly
// specified location. Use this only if you know what you are doing.
func UnsafeLocationFilter(location string) func(*SKU) bool {
	return func(s *SKU) bool {
		// TODO(ace): how to handle better?
		want, err := s.GetLocation()
		if err != nil {
			return false
		}
		return locationEquals(want, location)
	}
}

// MapFn is a convenience type for mapping.
type MapFn func(*SKU) SKU
