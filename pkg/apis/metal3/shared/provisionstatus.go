package shared

// ProvisionStatus holds the state information for a single target.
type ProvisionStatus struct {
	// An indiciator for what the provisioner is doing with the host.
	State ProvisioningState `json:"state"`

	// The machine's UUID from the underlying provisioning tool
	ID string `json:"ID"`

	// Image holds the details of the last image successfully
	// provisioned to the host.
	Image Image `json:"image,omitempty"`

	// The RootDevicehints set by the user
	RootDeviceHints *RootDeviceHints `json:"rootDeviceHints,omitempty"`
}
