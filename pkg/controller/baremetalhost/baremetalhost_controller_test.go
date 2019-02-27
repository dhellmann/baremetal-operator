package baremetalhost_test

import (
	goctx "context"
	"encoding/base64"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes/scheme"

	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	metalkubeapis "github.com/metalkube/baremetal-operator/pkg/apis"
	metalkubev1alpha1 "github.com/metalkube/baremetal-operator/pkg/apis/metalkube/v1alpha1"
	"github.com/metalkube/baremetal-operator/pkg/controller/baremetalhost"
	"github.com/metalkube/baremetal-operator/pkg/provisioning"
	"github.com/metalkube/baremetal-operator/pkg/utils"
)

var log = logf.Log.WithName("baremetalhost_controller_test")

const (
	namespace         string = "test-namespace"
	defaultSecretName string = "bmc-creds-valid"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
	// Register our package types with the global scheme
	metalkubeapis.AddToScheme(scheme.Scheme)
}

func makeSecret(name, username, password string) *corev1.Secret {
	data := make(map[string][]byte)
	data["username"] = []byte(base64.StdEncoding.EncodeToString([]byte(username)))
	data["password"] = []byte(base64.StdEncoding.EncodeToString([]byte(password)))

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: "1",
		},
		Data: data,
	}

	return secret
}

func makeHost(name string, spec *metalkubev1alpha1.BareMetalHostSpec) *metalkubev1alpha1.BareMetalHost {
	return &metalkubev1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BareMetalHost",
			APIVersion: "metalkube.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: *spec,
	}
}

func makeDefaultHost() *metalkubev1alpha1.BareMetalHost {
	spec := &metalkubev1alpha1.BareMetalHostSpec{
		BMC: metalkubev1alpha1.BMCDetails{
			IP:              "192.168.100.100",
			CredentialsName: defaultSecretName,
		},
	}
	return makeHost("test-host", spec)
}

func makeReconciler(initObjs ...runtime.Object) *baremetalhost.ReconcileBareMetalHost {

	c := fakeclient.NewFakeClient(initObjs...)

	// Add a default secret that can be used by most hosts.
	c.Create(goctx.TODO(), makeSecret(defaultSecretName, "User", "Pass"))

	return &baremetalhost.ReconcileBareMetalHost{
		Client: c,
		Scheme: scheme.Scheme,
		Provisioner: &provisioning.Provisioner{
			DeprovisionRequeueDelay: 0,
		},
	}
}

type DoneFunc func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool

func tryReconcile(t *testing.T, r *baremetalhost.ReconcileBareMetalHost, host *metalkubev1alpha1.BareMetalHost, isDone DoneFunc) {

	namespacedName := types.NamespacedName{
		Namespace: host.ObjectMeta.Namespace,
		Name:      host.ObjectMeta.Name,
	}
	request := reconcile.Request{NamespacedName: namespacedName}

	i := 0
	for {
		result, err := r.Reconcile(request)
		if err != nil {
			t.Fatal(err)
			break
		}

		i = i + 1
		if i >= 50 {
			t.Fatal(fmt.Errorf("Exceeded 50 iterations"))
		}

		// The FakeClient keeps a copy of the object we update, so we
		// need to replace the one we have with the updated data in
		// order to test it.
		r.Client.Get(goctx.TODO(), namespacedName, host)

		if isDone(host, result) {
			break
		}

		if !result.Requeue {
			t.Fatal(fmt.Errorf("Ended reconcile without test condition being true"))
			break
		}
	}
}

func waitForStatus(t *testing.T, r *baremetalhost.ReconcileBareMetalHost, host *metalkubev1alpha1.BareMetalHost, desiredStatus string) {
	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			state := host.OperationalStatus()
			t.Logf("OperationalState of %s: %s", host.ObjectMeta.Name, state)
			if state == desiredStatus {
				return true
			}
			return false
		},
	)
}

func waitForErrorStatus(t *testing.T, r *baremetalhost.ReconcileBareMetalHost, host *metalkubev1alpha1.BareMetalHost) {
	waitForStatus(t, r, host, metalkubev1alpha1.OperationalStatusError)
}

func waitForOfflineStatus(t *testing.T, r *baremetalhost.ReconcileBareMetalHost, host *metalkubev1alpha1.BareMetalHost) {
	waitForStatus(t, r, host, metalkubev1alpha1.OperationalStatusOffline)
}

func waitForOnlineStatus(t *testing.T, r *baremetalhost.ReconcileBareMetalHost, host *metalkubev1alpha1.BareMetalHost) {
	waitForStatus(t, r, host, metalkubev1alpha1.OperationalStatusOnline)
}

// TestAddFinalizers ensures that the finalizers for the host are
// updated as part of reconciling it.
func TestAddFinalizers(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("finalizers: %v", host.ObjectMeta.Finalizers)
			if utils.StringInList(host.ObjectMeta.Finalizers, metalkubev1alpha1.BareMetalHostFinalizer) {
				return true
			}
			return false
		},
	)
}

// TestSetLastUpdated ensures that the lastUpdated timestamp in the
// status is set to a non-zero value during reconciliation.
func TestSetLastUpdated(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("LastUpdated: %v", host.Status.LastUpdated)
			if !host.Status.LastUpdated.IsZero() {
				return true
			}
			return false
		},
	)
}

// TestUpdateCredentialsSecretSuccessFields ensures that the
// GoodCredentials fields are updated in the status block of a host
// when the secret used exists and has all of the right fields.
func TestUpdateCredentialsSecretSuccessFields(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
				host.Status.GoodCredentials.Version)
			if host.Status.GoodCredentials.Version != "" {
				return true
			}
			return false
		},
	)

}

// TestUpdateGoodCredentialsOnNewSecret ensures that the
// GoodCredentials fields are updated when the secret for a host is
// changed to another secret that is also good.
func TestUpdateGoodCredentialsOnNewSecret(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
				host.Status.GoodCredentials.Version)
			if host.Status.GoodCredentials.Version != "" {
				return true
			}
			return false
		},
	)

	// Define a second valid secret and update the host to use it.
	secret2 := makeSecret("bmc-creds-valid2", "User", "Pass")
	err := r.Client.Create(goctx.TODO(), secret2)
	if err != nil {
		t.Fatal(err)
	}

	host.Spec.BMC.CredentialsName = "bmc-creds-valid2"
	err = r.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
				host.Status.GoodCredentials.Version)
			if host.Status.GoodCredentials.Reference != nil && host.Status.GoodCredentials.Reference.Name == "bmc-creds-valid2" {
				return true
			}
			return false
		},
	)
}

// TestUpdateGoodCredentialsOnBadSecret ensures that the
// GoodCredentials fields are *not* updated when the secret is changed
// to one that is missing data.
func TestUpdateGoodCredentialsOnBadSecret(t *testing.T) {
	host := makeDefaultHost()
	badSecret := makeSecret("bmc-creds-no-user", "", "Pass")
	r := makeReconciler(host, badSecret)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
				host.Status.GoodCredentials.Version)
			if host.Status.GoodCredentials.Version != "" {
				return true
			}
			return false
		},
	)

	host.Spec.BMC.CredentialsName = "bmc-creds-no-user"
	err := r.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {

			t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
				host.Status.GoodCredentials.Version)
			if host.Spec.BMC.CredentialsName != "bmc-creds-no-user" {
				return false
			}
			if host.Status.GoodCredentials.Reference != nil && host.Status.GoodCredentials.Reference.Name == "bmc-creds-valid" {
				return true
			}
			return false
		},
	)
}

// TestMissingBMCParameters ensures that a host that is missing some
// of the required BMC settings is put into an error state.
func TestMissingBMCParameters(t *testing.T) {

	noIP := makeHost("missing-bmc-ip",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "",
				CredentialsName: "bmc-creds-valid",
			},
		})
	r := makeReconciler(noIP)
	waitForErrorStatus(t, r, noIP)

	secretNoUser := makeSecret("bmc-creds-no-user", "", "Pass")
	noUsername := makeHost("missing-bmc-username",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-user",
			},
		})
	r = makeReconciler(noUsername, secretNoUser)
	waitForErrorStatus(t, r, noUsername)

	secretNoPassword := makeSecret("bmc-creds-no-pass", "User", "")
	noPassword := makeHost("missing-bmc-password",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-pass",
			},
		})
	r = makeReconciler(noPassword, secretNoPassword)
	waitForErrorStatus(t, r, noPassword)
}

// TestFixSecret ensures that when the secret for a host is updated to
// be correct the status of the host moves out of the error state.
func TestFixSecret(t *testing.T) {

	secret := makeSecret("bmc-creds-no-user", "", "Pass")
	host := makeHost("fix-secret",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-user",
			},
		})
	r := makeReconciler(host, secret)
	waitForErrorStatus(t, r, host)

	secret = &corev1.Secret{}
	secretName := types.NamespacedName{
		Namespace: namespace,
		Name:      "bmc-creds-no-user",
	}
	err := r.Client.Get(goctx.TODO(), secretName, secret)
	if err != nil {
		t.Fatal(err)
	}
	secret.Data["username"] = []byte(base64.StdEncoding.EncodeToString([]byte("username")))
	err = r.Client.Update(goctx.TODO(), secret)
	if err != nil {
		t.Fatal(err)
	}
	waitForOfflineStatus(t, r, host)
}

// TestToggleOnline ensures that when the online flag changes the
// status of the host is updated.
func TestToggleOnline(t *testing.T) {
	host := makeHost("set-offline",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
			Online: true,
		})
	r := makeReconciler(host)
	waitForOnlineStatus(t, r, host)

	host.Spec.Online = false
	err := r.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	waitForOfflineStatus(t, r, host)

	host.Spec.Online = true
	err = r.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}
	waitForOnlineStatus(t, r, host)
}

// TestSetHardwareProfileLabel ensures that the host has a label with
// the hardware profile name.
func TestSetHardwareProfileLabel(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("labels: %v", host.ObjectMeta.Labels)
			if host.ObjectMeta.Labels[metalkubev1alpha1.HardwareProfileLabel] != "" {
				return true
			}
			return false
		},
	)
}

// TestCreateHardwareDetails ensures that the HardwareDetails portion
// of the status block is filled in for new hosts.
func TestCreateHardwareDetails(t *testing.T) {
	host := makeDefaultHost()
	r := makeReconciler(host)

	tryReconcile(t, r, host,
		func(host *metalkubev1alpha1.BareMetalHost, result reconcile.Result) bool {
			t.Logf("new host details: %v", host.Status.HardwareDetails)
			if host.Status.HardwareDetails != nil {
				return true
			}
			return false
		},
	)
}
