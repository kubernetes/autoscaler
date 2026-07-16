package skewer

// HasSCSISupport determines if a SKU supports SCSI disk controller type.
// If no disk controller types are declared, it assumes SCSI is supported for backward compatibility.
func (s *SKU) HasSCSISupport() bool {
	declaresSCSI := s.HasCapabilityWithSeparator(DiskControllerTypes, DiskControllerSCSI)
	declaresNothing := !(declaresSCSI || s.HasNVMeSupport())
	return declaresSCSI || declaresNothing
}

// HasNVMeSupport determines if a SKU supports NVMe disk controller type.
func (s *SKU) HasNVMeSupport() bool {
	return s.HasCapabilityWithSeparator(DiskControllerTypes, DiskControllerNVMe)
}

// SupportsNVMeEphemeralOSDisk determines if a SKU supports NVMe placement for ephemeral OS disk.
func (s *SKU) SupportsNVMeEphemeralOSDisk() bool {
	return s.HasCapabilityWithSeparator(SupportedEphemeralOSDiskPlacements, EphemeralDiskPlacementNvme)
}

// NVMeDiskSizeInMiB returns the NVMe disk size in MiB for the SKU.
// Returns an error if the capability is not found, nil, or cannot be parsed.
func (s *SKU) NVMeDiskSizeInMiB() (int64, error) {
	return s.GetCapabilityIntegerQuantity(NvmeDiskSizeInMiB)
}
