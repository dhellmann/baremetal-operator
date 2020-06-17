package shared

// ChecksumType holds the algorithm name for the checksum
// +kubebuilder:validation:Enum=md5;sha256;sha512
type ChecksumType string

const (
	// MD5 checksum type
	MD5 ChecksumType = "md5"

	// SHA256 checksum type
	SHA256 ChecksumType = "sha256"

	// SHA512 checksum type
	SHA512 ChecksumType = "sha512"
)
