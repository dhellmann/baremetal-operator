package shared

const (
	// BareMetalHostFinalizer is the name of the finalizer added to
	// hosts to block delete operations until the physical host can be
	// deprovisioned.
	BareMetalHostFinalizer string = "baremetalhost.metal3.io"

	// PausedAnnotation is the annotation that pauses the reconciliation (triggers
	// an immediate requeue)
	PausedAnnotation = "baremetalhost.metal3.io/paused"

	// StatusAnnotation is the annotation that keeps a copy of the Status of BMH
	// This is particularly useful when we pivot BMH. If the status
	// annotation is present and status is empty, BMO will reconstruct BMH Status
	// from the status annotation.
	StatusAnnotation = "baremetalhost.metal3.io/status"

	// UnhealthyAnnotation is the annotation that sets unhealthy status of BMH
	UnhealthyAnnotation = "baremetalhost.metal3.io/unhealthy"
)
