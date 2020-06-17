package shared

// ErrorType indicates the class of problem that has caused the Host resource
// to enter an error state.
// +kubebuilder:validation:Enum=registration error;inspection error;provisioning error;power management error
type ErrorType string

const (
	// RegistrationError is an error condition occurring when the
	// controller is unable to connect to the Host's baseboard management
	// controller.
	RegistrationError ErrorType = "registration error"
	// InspectionError is an error condition occurring when an attempt to
	// obtain hardware details from the Host fails.
	InspectionError ErrorType = "inspection error"
	// ProvisioningError is an error condition occuring when the controller
	// fails to provision or deprovision the Host.
	ProvisioningError ErrorType = "provisioning error"
	// PowerManagementError is an error condition occurring when the
	// controller is unable to modify the power state of the Host.
	PowerManagementError ErrorType = "power management error"
)
