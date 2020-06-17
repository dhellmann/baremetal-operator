package shared

// Image holds the details of an image either to provisioned or that
// has been provisioned.
type Image struct {
	// URL is a location of an image to deploy.
	URL string `json:"url"`

	// Checksum is the checksum for the image.
	Checksum string `json:"checksum"`

	// ChecksumType is the checksum algorithm for the image.
	// e.g md5, sha256, sha512
	ChecksumType ChecksumType `json:"checksumType,omitempty"`

	// DiskFormat contains the format of the image (raw, qcow2, ...)
	// Needs to be set to raw for raw images streaming
	// +kubebuilder:validation:Enum=raw;qcow2;vdi;vmdk
	DiskFormat *string `json:"format,omitempty"`
}
