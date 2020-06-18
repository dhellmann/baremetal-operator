package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
)

// NOTE: json tags are required.  Any new fields you add must have
// json tags for the fields to be serialized.

// NOTE(dhellmann): Update docs/api.md when changing these data structure.

// BareMetalHostSpec defines the desired state of BareMetalHost
type BareMetalHostSpec struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code
	// after modifying this file

	// Taints is the full, authoritative list of taints to apply to
	// the corresponding Machine. This list will overwrite any
	// modifications made to the Machine on an ongoing basis.
	// +optional
	Taints []corev1.Taint `json:"taints,omitempty"`

	// How do we connect to the BMC?
	BMC metal3shared.BMCDetails `json:"bmc,omitempty"`

	// What is the name of the hardware profile for this host? It
	// should only be necessary to set this when inspection cannot
	// automatically determine the profile.
	HardwareProfile string `json:"hardwareProfile,omitempty"`

	// Provide guidance about how to choose the device for the image
	// being provisioned.
	RootDeviceHints *metal3shared.RootDeviceHints `json:"rootDeviceHints,omitempty"`

	// Which MAC address will PXE boot? This is optional for some
	// types, but required for libvirt VMs driven by vbmc.
	// +kubebuilder:validation:Pattern=`[0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){5}`
	BootMACAddress string `json:"bootMACAddress,omitempty"`

	// Should the server be online?
	Online bool `json:"online"`

	// ConsumerRef can be used to store information about something
	// that is using a host. When it is not empty, the host is
	// considered "in use".
	ConsumerRef *corev1.ObjectReference `json:"consumerRef,omitempty"`

	// Image holds the details of the image to be provisioned.
	Image *metal3shared.Image `json:"image,omitempty"`

	// UserData holds the reference to the Secret containing the user
	// data to be passed to the host before it boots.
	UserData *corev1.SecretReference `json:"userData,omitempty"`

	// NetworkData holds the reference to the Secret containing network
	// configuration (e.g content of network_data.json which is passed
	// to Config Drive).
	NetworkData *corev1.SecretReference `json:"networkData,omitempty"`

	// MetaData holds the reference to the Secret containing host metadata
	// (e.g. meta_data.json which is passed to Config Drive).
	MetaData *corev1.SecretReference `json:"metaData,omitempty"`

	// Description is a human-entered text used to help identify the host
	Description string `json:"description,omitempty"`

	// ExternallyProvisioned means something else is managing the
	// image running on the host and the operator should only manage
	// the power status and hardware inventory inspection. If the
	// Image field is filled in, this field is ignored.
	ExternallyProvisioned bool `json:"externallyProvisioned,omitempty"`
}

// FIXME(dhellmann): We probably want some other module to own these
// data structures.

// ClockSpeed is a clock speed in MHz
type ClockSpeed float64

// ClockSpeed multipliers
const (
	MegaHertz ClockSpeed = 1.0
	GigaHertz            = 1000 * MegaHertz
)

// CPU describes one processor on the host.
type CPU struct {
	Arch           string     `json:"arch"`
	Model          string     `json:"model"`
	ClockMegahertz ClockSpeed `json:"clockMegahertz"`
	Flags          []string   `json:"flags"`
	Count          int        `json:"count"`
}

// HardwareDetails collects all of the information about hardware
// discovered on the host.
type HardwareDetails struct {
	SystemVendor metal3shared.HardwareSystemVendor `json:"systemVendor"`
	Firmware     metal3shared.Firmware             `json:"firmware"`
	RAMMebibytes int                               `json:"ramMebibytes"`
	NIC          []metal3shared.NIC                `json:"nics"`
	Storage      []metal3shared.Storage            `json:"storage"`
	CPU          CPU                               `json:"cpu"`
	Hostname     string                            `json:"hostname"`
}

// BareMetalHostStatus defines the observed state of BareMetalHost
type BareMetalHostStatus struct {
	// Important: Run "operator-sdk generate k8s" to regenerate code
	// after modifying this file

	// OperationalStatus holds the status of the host
	OperationalStatus metal3shared.OperationalStatus `json:"operationalStatus"`

	// ErrorType indicates the type of failure encountered when the
	// OperationalStatus is OperationalStatusError
	ErrorType metal3shared.ErrorType `json:"errorType,omitempty"`

	// LastUpdated identifies when this status was last observed.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// The name of the profile matching the hardware details.
	HardwareProfile string `json:"hardwareProfile"`

	// The hardware discovered to exist on the host.
	HardwareDetails *HardwareDetails `json:"hardware,omitempty"`

	// Information tracked by the provisioner.
	Provisioning metal3shared.ProvisionStatus `json:"provisioning"`

	// the last credentials we were able to validate as working
	GoodCredentials metal3shared.CredentialsStatus `json:"goodCredentials,omitempty"`

	// the last credentials we sent to the provisioning backend
	TriedCredentials metal3shared.CredentialsStatus `json:"triedCredentials,omitempty"`

	// the last error message reported by the provisioning subsystem
	ErrorMessage string `json:"errorMessage"`

	// indicator for whether or not the host is powered on
	PoweredOn bool `json:"poweredOn"`

	// OperationHistory holds information about operations performed
	// on this host.
	OperationHistory metal3shared.OperationHistory `json:"operationHistory"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BareMetalHost is the Schema for the baremetalhosts API
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=bmh;bmhost
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.operationalStatus",description="Operational status"
// +kubebuilder:printcolumn:name="Provisioning Status",type="string",JSONPath=".status.provisioning.state",description="Provisioning status"
// +kubebuilder:printcolumn:name="Consumer",type="string",JSONPath=".spec.consumerRef.name",description="Consumer using this host"
// +kubebuilder:printcolumn:name="BMC",type="string",JSONPath=".spec.bmc.address",description="Address of management controller"
// +kubebuilder:printcolumn:name="Hardware Profile",type="string",JSONPath=".status.hardwareProfile",description="The type of hardware detected"
// +kubebuilder:printcolumn:name="Online",type="string",JSONPath=".spec.online",description="Whether the host is online or not"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.errorMessage",description="Most recent error"
// +kubebuilder:storageversion
type BareMetalHost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BareMetalHostSpec   `json:"spec,omitempty"`
	Status BareMetalHostStatus `json:"status,omitempty"`
}

// Available returns true if the host is available to be provisioned.
func (host *BareMetalHost) Available() bool {
	if host.Spec.ConsumerRef != nil {
		return false
	}
	if host.GetDeletionTimestamp() != nil {
		return false
	}
	if host.HasError() {
		return false
	}
	return true
}

// SetErrorMessage updates the ErrorMessage in the host Status struct
// when necessary and returns true when a change is made or false when
// no change is made.
func (host *BareMetalHost) SetErrorMessage(errType metal3shared.ErrorType, message string) (dirty bool) {
	if host.Status.OperationalStatus != metal3shared.OperationalStatusError {
		host.Status.OperationalStatus = metal3shared.OperationalStatusError
		dirty = true
	}
	if host.Status.ErrorType != errType {
		host.Status.ErrorType = errType
		dirty = true
	}
	if host.Status.ErrorMessage != message {
		host.Status.ErrorMessage = message
		dirty = true
	}
	return dirty
}

// ClearError removes any existing error message.
func (host *BareMetalHost) ClearError() (dirty bool) {
	dirty = host.SetOperationalStatus(metal3shared.OperationalStatusOK)
	var emptyErrType metal3shared.ErrorType = ""
	if host.Status.ErrorType != emptyErrType {
		host.Status.ErrorType = emptyErrType
		dirty = true
	}
	if host.Status.ErrorMessage != "" {
		host.Status.ErrorMessage = ""
		dirty = true
	}
	return dirty
}

// setLabel updates the given label when necessary and returns true
// when a change is made or false when no change is made.
func (host *BareMetalHost) setLabel(name, value string) bool {
	if host.Labels == nil {
		host.Labels = make(map[string]string)
	}
	if host.Labels[name] != value {
		host.Labels[name] = value
		return true
	}
	return false
}

// getLabel returns the value associated with the given label. If
// there is no value, an empty string is returned.
func (host *BareMetalHost) getLabel(name string) string {
	if host.Labels == nil {
		return ""
	}
	return host.Labels[name]
}

// NeedsHardwareProfile returns true if the profile is not set
func (host *BareMetalHost) NeedsHardwareProfile() bool {
	return host.Status.HardwareProfile == ""
}

// HardwareProfile returns the hardware profile name for the host.
func (host *BareMetalHost) HardwareProfile() string {
	return host.Status.HardwareProfile
}

// SetHardwareProfile updates the hardware profile name and returns
// true when a change is made or false when no change is made.
func (host *BareMetalHost) SetHardwareProfile(name string) (dirty bool) {
	if host.Status.HardwareProfile != name {
		host.Status.HardwareProfile = name
		dirty = true
	}
	return dirty
}

// SetOperationalStatus updates the OperationalStatus field and returns
// true when a change is made or false when no change is made.
func (host *BareMetalHost) SetOperationalStatus(status metal3shared.OperationalStatus) bool {
	if host.Status.OperationalStatus != status {
		host.Status.OperationalStatus = status
		return true
	}
	return false
}

// OperationalStatus returns the contents of the OperationalStatus
// field.
func (host *BareMetalHost) OperationalStatus() metal3shared.OperationalStatus {
	return host.Status.OperationalStatus
}

// HasError returns a boolean indicating whether there is an error
// set for the host.
func (host *BareMetalHost) HasError() bool {
	return host.Status.ErrorMessage != ""
}

// CredentialsKey returns a NamespacedName suitable for loading the
// Secret containing the credentials associated with the host.
func (host *BareMetalHost) CredentialsKey() types.NamespacedName {
	return types.NamespacedName{
		Name:      host.Spec.BMC.CredentialsName,
		Namespace: host.ObjectMeta.Namespace,
	}
}

// NeedsHardwareInspection looks at the state of the host to determine
// if hardware inspection should be run.
func (host *BareMetalHost) NeedsHardwareInspection() bool {
	if host.Spec.ExternallyProvisioned {
		// Never perform inspection if we already know something is
		// using the host and we didn't provision it.
		return false
	}
	if host.WasProvisioned() {
		// Never perform inspection if we have already provisioned
		// this host, because we don't want to reboot it.
		return false
	}
	return host.Status.HardwareDetails == nil
}

// NeedsProvisioning compares the settings with the provisioning
// status and returns true when more work is needed or false
// otherwise.
func (host *BareMetalHost) NeedsProvisioning() bool {
	if !host.Spec.Online {
		// The host is not supposed to be powered on.
		return false
	}
	if host.Spec.Image == nil {
		// Without an image, there is nothing to provision.
		return false
	}
	if host.Spec.Image.URL == "" {
		// We have an Image struct but it is empty
		return false
	}
	if host.Status.Provisioning.Image.URL == "" {
		// We have an image set, but not provisioned.
		return true
	}
	return false
}

// WasProvisioned returns true when we think we have placed an image
// on the host.
func (host *BareMetalHost) WasProvisioned() bool {
	if host.Spec.ExternallyProvisioned {
		return false
	}
	if host.Status.Provisioning.Image.URL != "" {
		// We have an image provisioned.
		return true
	}
	return false
}

// NeedsDeprovisioning compares the settings with the provisioning
// status and returns true when the host should be deprovisioned.
func (host *BareMetalHost) NeedsDeprovisioning() bool {
	if host.Spec.Image == nil {
		return true
	}
	if host.Spec.Image.URL == "" {
		return true
	}
	if host.Status.Provisioning.Image.URL == "" {
		return false
	}
	if host.Spec.Image.URL != host.Status.Provisioning.Image.URL {
		return true
	}
	return false
}

// UpdateGoodCredentials modifies the GoodCredentials portion of the
// Status struct to record the details of the secret containing
// credentials known to work.
func (host *BareMetalHost) UpdateGoodCredentials(currentSecret corev1.Secret) {
	host.Status.GoodCredentials.Version = currentSecret.ObjectMeta.ResourceVersion
	host.Status.GoodCredentials.Reference = &corev1.SecretReference{
		Name:      currentSecret.ObjectMeta.Name,
		Namespace: currentSecret.ObjectMeta.Namespace,
	}
}

// UpdateTriedCredentials modifies the TriedCredentials portion of the
// Status struct to record the details of the secret containing
// credentials known to work.
func (host *BareMetalHost) UpdateTriedCredentials(currentSecret corev1.Secret) {
	host.Status.TriedCredentials.Version = currentSecret.ObjectMeta.ResourceVersion
	host.Status.TriedCredentials.Reference = &corev1.SecretReference{
		Name:      currentSecret.ObjectMeta.Name,
		Namespace: currentSecret.ObjectMeta.Namespace,
	}
}

// NewEvent creates a new event associated with the object and ready
// to be published to the kubernetes API.
func (host *BareMetalHost) NewEvent(reason, message string) corev1.Event {
	t := metav1.Now()
	return corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: reason + "-",
			Namespace:    host.ObjectMeta.Namespace,
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       "BareMetalHost",
			Namespace:  host.Namespace,
			Name:       host.Name,
			UID:        host.UID,
			APIVersion: SchemeGroupVersion.String(),
		},
		Reason:  reason,
		Message: message,
		Source: corev1.EventSource{
			Component: "metal3-baremetal-controller",
		},
		FirstTimestamp:      t,
		LastTimestamp:       t,
		Count:               1,
		Type:                corev1.EventTypeNormal,
		ReportingController: "metal3.io/baremetal-controller",
		Related:             host.Spec.ConsumerRef,
	}
}

// OperationMetricForState returns a pointer to the metric for the given
// provisioning state.
func (host *BareMetalHost) OperationMetricForState(operation metal3shared.ProvisioningState) (metric *metal3shared.OperationMetric) {
	history := &host.Status.OperationHistory
	switch operation {
	case metal3shared.StateRegistering:
		metric = &history.Register
	case metal3shared.StateInspecting:
		metric = &history.Inspect
	case metal3shared.StateProvisioning:
		metric = &history.Provision
	case metal3shared.StateDeprovisioning:
		metric = &history.Deprovision
	}
	return
}

// GetImageChecksum returns the hash value and its algo.
func (host *BareMetalHost) GetImageChecksum() (string, string, bool) {
	if host.Spec.Image == nil {
		return "", "", false
	}

	checksum := host.Spec.Image.Checksum
	checksumType := host.Spec.Image.ChecksumType

	if checksum == "" {
		// Return empty if checksum is not provided
		return "", "", false
	}
	if checksumType == "" {
		// If only checksum is specified. Assume type is md5
		return checksum, string(metal3shared.MD5), true
	}
	switch checksumType {
	case metal3shared.MD5, metal3shared.SHA256, metal3shared.SHA512:
		return checksum, string(checksumType), true
	default:
		return "", "", false
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BareMetalHostList contains a list of BareMetalHost
type BareMetalHostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BareMetalHost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BareMetalHost{}, &BareMetalHostList{})
}
