package skewer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/pkg/errors"
)

// SKU wraps an Azure compute SKU with richer functionality
type SKU compute.ResourceSku

// ErrCapabilityNotFound will be returned when a capability could not be
// found, even without a value.
type ErrCapabilityNotFound struct {
	capability string
}

func (e *ErrCapabilityNotFound) Error() string {
	return e.capability + "CapabilityNotFound"
}

// ErrCapabilityValueNil will be returned when a capability was found by
// name but the value was nil.
type ErrCapabilityValueNil struct {
	capability string
}

func (e *ErrCapabilityValueNil) Error() string {
	return e.capability + "CapabilityValueNil"
}

// ErrCapabilityValueParse will be returned when a capability was found by
// name but the value was nil.
type ErrCapabilityValueParse struct {
	capability string
	value      string
	err        error
}

func (e *ErrCapabilityValueParse) Error() string {
	return fmt.Sprintf("%sCapabilityValueParse: failed to parse string '%s' as int64, error: '%s'", e.capability, e.value, e.err)
}

// VCPU returns the number of vCPUs this SKU supports.
func (s *SKU) VCPU() (int64, error) {
	return s.GetCapabilityIntegerQuantity(VCPUs)
}

// Memory returns the amount of memory this SKU supports.
func (s *SKU) Memory() (float64, error) {
	return s.GetCapabilityFloatQuantity(MemoryGB)
}

// MaxCachedDiskBytes returns the number of bytes available for the
// cache if it exists on this VM size.
func (s *SKU) MaxCachedDiskBytes() (int64, error) {
	return s.GetCapabilityIntegerQuantity(CachedDiskBytes)
}

// MaxResourceVolumeMB returns the number of bytes available for the
// cache if it exists on this VM size.
func (s *SKU) MaxResourceVolumeMB() (int64, error) {
	return s.GetCapabilityIntegerQuantity(MaxResourceVolumeMB)
}

// IsEncryptionAtHostSupported returns true when Encryption at Host is
// supported for the VM size.
func (s *SKU) IsEncryptionAtHostSupported() bool {
	return s.HasCapability(EncryptionAtHost)
}

// From ultra SSD documentation
//   https://docs.microsoft.com/en-us/azure/virtual-machines/disks-enable-ultra-ssd
// Ultra SSD can be either supported on
// 		- "Single VMs" without availability zone support, or
// 		- On availability zones
// So provide functions to test both cases

// IsUltraSSDAvailableWithoutAvailabilityZone returns true when a VM size has ultra SSD enabled
// in the region
func (s *SKU) IsUltraSSDAvailableWithoutAvailabilityZone() bool {
	return s.HasCapability(UltraSSDAvailable)
}

// IsUltraSSDAvailableInAvailabilityZone returns true when a VM size has ultra SSD enabled
// in the given availability zone
func (s *SKU) IsUltraSSDAvailableInAvailabilityZone(zone string) bool {
	return s.HasCapabilityInZone(UltraSSDAvailable, zone)
}

// IsUltraSSDAvailable returns true when a VM size has ultra SSD enabled
// in at least 1 unrestricted zone.
//
// Deprecated. Use either IsUltraSSDAvailableWithoutAvailabilityZone or IsUltraSSDAvailableInAvailabilityZone
func (s *SKU) IsUltraSSDAvailable() bool {
	return s.HasZonalCapability(UltraSSDAvailable)
}

// IsEphemeralOSDiskSupported returns true when the VM size supports
// ephemeral OS.
func (s *SKU) IsEphemeralOSDiskSupported() bool {
	return s.HasCapability(EphemeralOSDisk)
}

// IsAcceleratedNetworkingSupported returns true when the VM size supports
// accelerated networking.
func (s *SKU) IsAcceleratedNetworkingSupported() bool {
	return s.HasCapability(AcceleratedNetworking)
}

// IsHyperVGen2Supported returns true when the VM size supports
// accelerated networking.
func (s *SKU) IsHyperVGen2Supported() bool {
	return s.HasCapabilityWithSeparator(HyperVGenerations, HyperVGeneration2)
}

// GetCapabilityIntegerQuantity retrieves and parses the value of an
// integer numeric capability with the provided name. It errors if the
// capability is not found, the value was nil, or the value could not be
// parsed as an integer.
func (s *SKU) GetCapabilityIntegerQuantity(name string) (int64, error) {
	if s.Capabilities == nil {
		return -1, &ErrCapabilityNotFound{name}
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && *capability.Name == name {
			if capability.Value != nil {
				intVal, err := strconv.ParseInt(*capability.Value, 10, 64)
				if err != nil {
					return -1, &ErrCapabilityValueParse{name, *capability.Value, err}
				}
				return intVal, nil
			}
			return -1, &ErrCapabilityValueNil{name}
		}
	}
	return -1, &ErrCapabilityNotFound{name}
}

// GetCapabilityFloatQuantity retrieves and parses the value of a
// floating point numeric capability with the provided name. It errors
// if the capability is not found, the value was nil, or the value could
// not be parsed as an integer.
func (s *SKU) GetCapabilityFloatQuantity(name string) (float64, error) {
	if s.Capabilities == nil {
		return -1, &ErrCapabilityNotFound{name}
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && *capability.Name == name {
			if capability.Value != nil {
				intVal, err := strconv.ParseFloat(*capability.Value, 64)
				if err != nil {
					return -1, &ErrCapabilityValueParse{name, *capability.Value, err}
				}
				return intVal, nil
			}
			return -1, &ErrCapabilityValueNil{name}
		}
	}
	return -1, &ErrCapabilityNotFound{name}
}

// HasCapability return true for a capability which can be either
// supported or not. Examples include "EphemeralOSDiskSupported",
// "EncryptionAtHostSupported", "AcceleratedNetworkingEnabled", and
// "RdmaEnabled"
func (s *SKU) HasCapability(name string) bool {
	if s.Capabilities == nil {
		return false
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			return capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported))
		}
	}
	return false
}

// HasZonalCapability return true for a capability which can be either
// supported or not. Examples include "UltraSSDAvailable".
// This function only checks that zone details suggest support: it will
// return true for a whole location even when only one zone supports the
// feature. Currently, the only real scenario that appears to use
// zoneDetails is UltraSSDAvailable which always lists all regions as
// available.
// For per zone capability check, use "HasCapabilityInZone"
func (s *SKU) HasZonalCapability(name string) bool {
	if s.LocationInfo == nil {
		return false
	}
	for _, locationInfo := range *s.LocationInfo {
		if locationInfo.ZoneDetails == nil {
			continue
		}
		for _, zoneDetails := range *locationInfo.ZoneDetails {
			if zoneDetails.Capabilities == nil {
				continue
			}
			for _, capability := range *zoneDetails.Capabilities {
				if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
					if capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported)) {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasCapabilityInZone return true if the specified capability name is supported in the
// specified zone.
func (s *SKU) HasCapabilityInZone(name string, zone string) bool {
	if s.LocationInfo == nil {
		return false
	}
	for _, locationInfo := range *s.LocationInfo {
		if locationInfo.ZoneDetails == nil {
			continue
		}
		for _, zoneDetails := range *locationInfo.ZoneDetails {
			if zoneDetails.Capabilities == nil {
				continue
			}
			foundZone := false
			if zoneDetails.Name != nil {
				for _, zoneName := range *zoneDetails.Name {
					if strings.EqualFold(zone, zoneName) {
						foundZone = true
						break
					}
				}
			}
			if !foundZone {
				continue
			}

			for _, capability := range *zoneDetails.Capabilities {
				if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
					if capability.Value != nil && strings.EqualFold(*capability.Value, string(CapabilitySupported)) {
						return true
					}
				}
			}
		}
	}
	return false
}

// HasCapabilityWithSeparator return true for a capability which may be
// exposed as a comma-separated list. We check that the list contains
// the desired substring. An example is "HyperVGenerations" which may be
// "V1,V2"
func (s *SKU) HasCapabilityWithSeparator(name, value string) bool {
	if s.Capabilities == nil {
		return false
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			return capability.Value != nil && strings.Contains(normalizeLocation(*capability.Value), normalizeLocation(value))
		}
	}
	return false
}

// HasCapabilityWithMinCapacity returns true when the SKU has a
// capability with the requested name, and the value is greater than or
// equal to the desired value.
// "MaxResourceVolumeMB", "OSVhdSizeMB", "vCPUs",
// "MemoryGB","MaxDataDiskCount", "CombinedTempDiskAndCachedIOPS",
// "CombinedTempDiskAndCachedReadBytesPerSecond",
// "CombinedTempDiskAndCachedWriteBytesPerSecond", "UncachedDiskIOPS",
// and "UncachedDiskBytesPerSecond"
func (s *SKU) HasCapabilityWithMinCapacity(name string, value int64) (bool, error) {
	if s.Capabilities == nil {
		return false, nil
	}
	for _, capability := range *s.Capabilities {
		if capability.Name != nil && strings.EqualFold(*capability.Name, name) {
			if capability.Value != nil {
				intVal, err := strconv.ParseInt(*capability.Value, 10, 64)
				if err != nil {
					return false, errors.Wrapf(err, "failed to parse string '%s' as int64", *capability.Value)
				}
				if intVal >= value {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}

// IsAvailable returns true when the requested location matches one on
// the sku, and there are no total restrictions on the location.
func (s *SKU) IsAvailable(location string) bool {
	if s.LocationInfo == nil {
		return false
	}
	for _, locationInfo := range *s.LocationInfo {
		if locationInfo.Location != nil {
			if locationEquals(*locationInfo.Location, location) {
				if s.Restrictions != nil {
					for _, restriction := range *s.Restrictions {
						// Can't deploy to any zones in this location. We're done.
						if restriction.Type == compute.Location {
							return false
						}
					}
				}
				return true
			}
		}
	}
	return false
}

// IsRestricted returns true when a location restriction exists for
// this SKU.
func (s *SKU) IsRestricted(location string) bool {
	if s.Restrictions == nil {
		return false
	}
	for _, restriction := range *s.Restrictions {
		if restriction.Values == nil {
			continue
		}
		for _, candidate := range *restriction.Values {
			// Can't deploy in this location. We're done.
			if locationEquals(candidate, location) && restriction.Type == compute.Location {
				return true
			}
		}
	}
	return false
}

// IsResourceType returns true when the wrapped SKU has the provided
// value as its resource type. This may be used to filter using values
// such as "virtualMachines", "disks", "availabilitySets", "snapshots",
// and "hostGroups/hosts".
func (s *SKU) IsResourceType(t string) bool {
	return s.ResourceType != nil && strings.EqualFold(*s.ResourceType, t)
}

// GetResourceType returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "virtualMachines" for a virtual machine.
func (s *SKU) GetResourceType() string {
	if s.ResourceType == nil {
		return ""
	}
	return *s.ResourceType
}

// GetName returns the name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "Standard_D8s_v3" for a virtual machine.
func (s *SKU) GetName() string {
	if s.Name == nil {
		return ""
	}

	return *s.Name
}

// GetFamilyName returns the family name of this resource sku. It normalizes pointers
// to the empty string for comparison purposes. For example,
// "standardDSv2Family" for a virtual machine.
func (s *SKU) GetFamilyName() string {
	if s.Family == nil {
		return ""
	}

	return *s.Family
}

// GetLocation returns the location for a given SKU.
func (s *SKU) GetLocation() (string, error) {
	if s.Locations == nil {
		return "", fmt.Errorf("sku had nil location array")
	}

	if len(*s.Locations) < 1 {
		return "", fmt.Errorf("sku had no locations")
	}

	if len(*s.Locations) > 1 {
		return "", fmt.Errorf("sku had multiple locations, refusing to disambiguate")
	}

	return (*s.Locations)[0], nil
}

// HasLocation returns true if the given sku exposes this region for deployment.
func (s *SKU) HasLocation(location string) bool {
	if s.Locations == nil {
		return false
	}

	for _, candidate := range *s.Locations {
		if locationEquals(candidate, location) {
			return true
		}
	}

	return false
}

// HasLocationRestriction returns true if the location is restricted for
// this sku.
func (s *SKU) HasLocationRestriction(location string) bool {
	if s.Restrictions == nil {
		return false
	}

	for _, restriction := range *s.Restrictions {
		if restriction.Type != compute.Location {
			continue
		}
		if restriction.Values == nil {
			continue
		}
		for _, candidate := range *restriction.Values {
			if locationEquals(candidate, location) {
				return true
			}
		}
	}

	return false
}

// AvailabilityZones returns the list of Availability Zones which have this resource SKU available and unrestricted.
func (s *SKU) AvailabilityZones(location string) map[string]bool { // nolint:gocyclo
	if s.LocationInfo == nil {
		return nil
	}

	// Use map for easy deletion and iteration
	availableZones := make(map[string]bool)
	restrictedZones := make(map[string]bool)

	for _, locationInfo := range *s.LocationInfo {
		if locationInfo.Location == nil {
			continue
		}
		if locationEquals(*locationInfo.Location, location) {
			// add all zones
			if locationInfo.Zones != nil {
				for _, zone := range *locationInfo.Zones {
					availableZones[zone] = true
				}
			}

			// iterate restrictions, remove any restricted zones for this location
			if s.Restrictions != nil {
				for _, restriction := range *s.Restrictions {
					if restriction.Values != nil {
						for _, candidate := range *restriction.Values {
							if locationEquals(candidate, location) {
								if restriction.Type == compute.Location {
									// Can't deploy in this location. We're done.
									return nil
								}

								if restriction.RestrictionInfo != nil && restriction.RestrictionInfo.Zones != nil {
									// remove restricted zones
									for _, zone := range *restriction.RestrictionInfo.Zones {
										restrictedZones[zone] = true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for zone := range restrictedZones {
		delete(availableZones, zone)
	}

	return availableZones
}

// Equal returns true when two skus have the same location, type, and name.
func (s *SKU) Equal(other *SKU) bool {
	location, localErr := s.GetLocation()
	otherLocation, otherErr := s.GetLocation()
	return strings.EqualFold(s.GetResourceType(), other.GetResourceType()) &&
		strings.EqualFold(s.GetName(), other.GetName()) &&
		locationEquals(location, otherLocation) &&
		localErr != nil &&
		otherErr != nil
}
