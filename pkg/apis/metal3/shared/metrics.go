package shared

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OperationMetric contains metadata about an operation (inspection,
// provisioning, etc.) used for tracking metrics.
type OperationMetric struct {
	// +nullable
	Start metav1.Time `json:"start,omitempty"`
	// +nullable
	End metav1.Time `json:"end,omitempty"`
}

// Duration returns the length of time that was spent on the
// operation. If the operation is not finished, it returns 0.
func (om OperationMetric) Duration() time.Duration {
	if om.Start.IsZero() {
		return 0
	}
	return om.End.Time.Sub(om.Start.Time)
}

// OperationHistory holds information about operations performed on a
// host.
type OperationHistory struct {
	Register    OperationMetric `json:"register,omitempty"`
	Inspect     OperationMetric `json:"inspect,omitempty"`
	Provision   OperationMetric `json:"provision,omitempty"`
	Deprovision OperationMetric `json:"deprovision,omitempty"`
}
