package baremetalhost

import (
	goctx "context"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes/scheme"

	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	metal3apis "github.com/metal3-io/baremetal-operator/pkg/apis"
	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
	metal3 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha2"
	"github.com/metal3-io/baremetal-operator/pkg/provisioner/demo"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
	// Register our package types with the global scheme
	metal3apis.AddToScheme(scheme.Scheme)
}

func newDemoReconciler(initObjs ...runtime.Object) *ReconcileBareMetalHost {

	c := fakeclient.NewFakeClient(initObjs...)

	// Add a default secret that can be used by most hosts.
	bmcSecret := newSecret(defaultSecretName, map[string]string{"username": "User", "password": "Pass"})
	c.Create(goctx.TODO(), bmcSecret)

	return &ReconcileBareMetalHost{
		client:             c,
		scheme:             scheme.Scheme,
		provisionerFactory: demo.New,
	}
}

// TestDemoRegistrationError tests that a host with the right name reports
// a registration error
func TestDemoRegistrationError(t *testing.T) {
	host := newDefaultNamedHost(demo.RegistrationErrorHost, t)
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.HasError()
		},
	)
}

// TestDemoRegistering tests that a host with the right name reports
// that it is being registered
func TestDemoRegistering(t *testing.T) {
	host := newDefaultNamedHost(demo.RegisteringHost, t)
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.Status.Provisioning.State == metal3shared.StateRegistering
		},
	)
}

// TestDemoInspecting tests that a host with the right name reports
// that it is being inspected
func TestDemoInspecting(t *testing.T) {
	host := newDefaultNamedHost(demo.InspectingHost, t)
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.Status.Provisioning.State == metal3shared.StateInspecting
		},
	)
}

// TestDemoReady tests that a host with the right name reports
// that it is ready to be provisioned
func TestDemoReady(t *testing.T) {
	host := newDefaultNamedHost(demo.ReadyHost, t)
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.Status.Provisioning.State == metal3shared.StateReady
		},
	)
}

// TestDemoProvisioning tests that a host with the right name reports
// that it is being provisioned
func TestDemoProvisioning(t *testing.T) {
	host := newDefaultNamedHost(demo.ProvisioningHost, t)
	host.Spec.Image = &metal3.Image{
		URL:      "a-url",
		Checksum: "a-checksum",
	}
	host.Spec.Online = true
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.Status.Provisioning.State == metal3shared.StateProvisioning
		},
	)
}

// TestDemoProvisioned tests that a host with the right name reports
// that it has been provisioned
func TestDemoProvisioned(t *testing.T) {
	host := newDefaultNamedHost(demo.ProvisionedHost, t)
	host.Spec.Image = &metal3.Image{
		URL:      "a-url",
		Checksum: "a-checksum",
	}
	host.Spec.Online = true
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.Status.Provisioning.State == metal3shared.StateProvisioned
		},
	)
}

// TestDemoValidationError tests that a host with the right name
// reports that it had and error while being provisioned
func TestDemoValidationError(t *testing.T) {
	host := newDefaultNamedHost(demo.ValidationErrorHost, t)
	host.Spec.Image = &metal3.Image{
		URL:      "a-url",
		Checksum: "a-checksum",
	}
	host.Spec.Online = true
	r := newDemoReconciler(host)

	tryReconcile(t, r, host,
		func(host *metal3.BareMetalHost, result reconcile.Result) bool {
			t.Logf("Status: %q State: %q ErrorMessage: %q",
				host.OperationalStatus(),
				host.Status.Provisioning.State,
				host.Status.ErrorMessage,
			)
			return host.HasError()
		},
	)
}
