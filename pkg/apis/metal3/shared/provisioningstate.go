package shared

// ProvisioningState defines the states the provisioner will report
// the host has having.
type ProvisioningState string

const (
	// StateNone means the state is unknown
	StateNone ProvisioningState = ""

	// StateRegistrationError means there was an error registering the
	// host with the backend
	StateRegistrationError ProvisioningState = "registration error"

	// StateRegistering means we are telling the backend about the host
	StateRegistering ProvisioningState = "registering"

	// StateMatchProfile means we are comparing the discovered details
	// against known hardware profiles
	StateMatchProfile ProvisioningState = "match profile"

	// StateReady means the host can be consumed
	StateReady ProvisioningState = "ready"

	// StateAvailable means the host can be consumed
	StateAvailable ProvisioningState = "available"

	// StateProvisioning means we are writing an image to the host's
	// disk(s)
	StateProvisioning ProvisioningState = "provisioning"

	// StateProvisioningError means we are writing an image to the
	// host's disk(s)
	StateProvisioningError ProvisioningState = "provisioning error"

	// StateProvisioned means we have written an image to the host's
	// disk(s)
	StateProvisioned ProvisioningState = "provisioned"

	// StateExternallyProvisioned means something else is managing the
	// image on the host
	StateExternallyProvisioned ProvisioningState = "externally provisioned"

	// StateDeprovisioning means we are removing an image from the
	// host's disk(s)
	StateDeprovisioning ProvisioningState = "deprovisioning"

	// StateInspecting means we are running the agent on the host to
	// learn about the hardware components available there
	StateInspecting ProvisioningState = "inspecting"

	// StatePowerManagementError means something went wrong trying to
	// power the server on or off.
	StatePowerManagementError ProvisioningState = "power management error"

	// StateDeleting means we are in the process of cleaning up the host
	// ready for deletion
	StateDeleting ProvisioningState = "deleting"
)
