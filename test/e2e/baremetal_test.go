// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	goctx "context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	apis "github.com/metalkube/baremetal-operator/pkg/apis"
	metalkubev1alpha1 "github.com/metalkube/baremetal-operator/pkg/apis/metalkube/v1alpha1"
	"github.com/metalkube/baremetal-operator/pkg/utils"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 10
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

// Set up the test system to know about our types and return a
// context.
func setup(t *testing.T) *framework.TestCtx {
	bmhList := &metalkubev1alpha1.BareMetalHostList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BareMetalHost",
			APIVersion: "baremetalhosts.metalkube.org/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, bmhList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	t.Parallel()
	ctx := framework.NewTestCtx(t)

	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	makeSecret(t, ctx, "bmc-creds-valid", "User", "Pass")
	makeSecret(t, ctx, "bmc-creds-no-user", "", "Pass")
	makeSecret(t, ctx, "bmc-creds-no-pass", "User", "")

	return ctx
}

// Create a new BareMetalHost instance.
func newHost(t *testing.T, ctx *framework.TestCtx, name string, spec *metalkubev1alpha1.BareMetalHostSpec) *metalkubev1alpha1.BareMetalHost {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Using namespace: %v\n", namespace)

	host := &metalkubev1alpha1.BareMetalHost{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BareMetalHost",
			APIVersion: "baremetalhosts.metalkubev1alpha1.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", ctx.GetID(), name),
			Namespace: namespace,
		},
		Spec: *spec,
	}

	return host
}

// Create a BareMetalHost and publish it to the test system.
func makeHost(t *testing.T, ctx *framework.TestCtx, name string, spec *metalkubev1alpha1.BareMetalHostSpec) *metalkubev1alpha1.BareMetalHost {
	host := newHost(t, ctx, name, spec)

	// get global framework variables
	f := framework.Global

	// use TestCtx's create helper to create the object and add a
	// cleanup function for the new object
	err := f.Client.Create(
		goctx.TODO(),
		host,
		&framework.CleanupOptions{
			TestContext:   ctx,
			Timeout:       cleanupTimeout,
			RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	return host
}

func makeSecret(t *testing.T, ctx *framework.TestCtx, name string, username string, password string) {

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	data := make(map[string][]byte)
	data["username"] = []byte(base64.StdEncoding.EncodeToString([]byte(username)))
	data["password"] = []byte(base64.StdEncoding.EncodeToString([]byte(password)))

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	f := framework.Global
	err = f.Client.Create(
		goctx.TODO(),
		secret,
		&framework.CleanupOptions{
			TestContext:   ctx,
			Timeout:       cleanupTimeout,
			RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

}

type DoneFunc func(host *metalkubev1alpha1.BareMetalHost) (bool, error)

func refreshHost(host *metalkubev1alpha1.BareMetalHost) error {
	f := framework.Global
	namespacedName := types.NamespacedName{
		Namespace: host.ObjectMeta.Namespace,
		Name:      host.ObjectMeta.Name,
	}
	return f.Client.Get(goctx.TODO(), namespacedName, host)
}

func waitForHostStateChange(t *testing.T, host *metalkubev1alpha1.BareMetalHost, isDone DoneFunc) *metalkubev1alpha1.BareMetalHost {
	instance := &metalkubev1alpha1.BareMetalHost{}
	instance.ObjectMeta = host.ObjectMeta

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		t.Log("polling host for updates")
		refreshHost(instance)
		if err != nil {
			return false, err
		}
		done, err = isDone(instance)
		return done, err
	})
	if err != nil {
		t.Fatal(err)
	}

	return instance
}

func waitForOfflineStatus(t *testing.T, host *metalkubev1alpha1.BareMetalHost) {
	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState: %s", state)
		if state == metalkubev1alpha1.OperationalStatusOffline {
			return true, nil
		}
		return false, nil
	})
}

func waitForErrorStatus(t *testing.T, host *metalkubev1alpha1.BareMetalHost) {
	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState: %s", state)
		if state == metalkubev1alpha1.OperationalStatusError {
			return true, nil
		}
		return false, nil
	})
}

func TestAddFinalizers(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "gets-finalizers",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("finalizers: %v", host.ObjectMeta.Finalizers)
		if utils.StringInList(host.ObjectMeta.Finalizers, metalkubev1alpha1.BareMetalHostFinalizer) {
			return true, nil
		}
		return false, nil
	})
}

func TestUpdateCredentialsSecretSuccessFields(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "updates-success",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
			host.Status.GoodCredentials.Version)
		if host.Status.GoodCredentials.Version != "" {
			return true, nil
		}
		return false, nil
	})
}

func TestUpdateGoodCredentialsOnNewSecret(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()
	f := framework.Global

	host := makeHost(t, ctx, "updates-success",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
			host.Status.GoodCredentials.Version)
		if host.Status.GoodCredentials.Version != "" {
			return true, nil
		}
		return false, nil
	})

	makeSecret(t, ctx, "bmc-creds-valid2", "User", "Pass")

	refreshHost(host)
	host.Spec.BMC.CredentialsName = "bmc-creds-valid2"
	err := f.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
			host.Status.GoodCredentials.Version)
		if host.Status.GoodCredentials.Reference != nil && host.Status.GoodCredentials.Reference.Name == "bmc-creds-valid2" {
			return true, nil
		}
		return false, nil
	})
}

func TestUpdateGoodCredentialsOnBadSecret(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()
	f := framework.Global

	host := makeHost(t, ctx, "updates-success",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
			host.Status.GoodCredentials.Version)
		if host.Status.GoodCredentials.Version != "" {
			return true, nil
		}
		return false, nil
	})

	refreshHost(host)
	host.Spec.BMC.CredentialsName = "bmc-creds-no-user"
	err := f.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("ref: %v ver: %s", host.Status.GoodCredentials.Reference,
			host.Status.GoodCredentials.Version)
		if host.Spec.BMC.CredentialsName != "bmc-creds-no-user" {
			return false, nil
		}
		if host.Status.GoodCredentials.Reference != nil && host.Status.GoodCredentials.Reference.Name == "bmc-creds-valid" {
			return true, nil
		}
		return false, nil
	})
}

func TestSetLastUpdated(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "gets-last-updated",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("LastUpdated: %v", host.Status.LastUpdated)
		if !host.Status.LastUpdated.IsZero() {
			return true, nil
		}
		return false, nil
	})
}

func TestMissingBMCParameters(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	noIP := makeHost(t, ctx, "missing-bmc-ip",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "",
				CredentialsName: "bmc-creds-valid",
			},
		})
	waitForErrorStatus(t, noIP)

	noUsername := makeHost(t, ctx, "missing-bmc-username",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-user",
			},
		})
	waitForErrorStatus(t, noUsername)

	noPassword := makeHost(t, ctx, "missing-bmc-password",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-pass",
			},
		})
	waitForErrorStatus(t, noPassword)
}

func TestChangeSecret(t *testing.T) {
	// Create the host using the secret that does not have a username,
	// then modify the secret and look for the host status to change.

	ctx := setup(t)
	defer ctx.Cleanup()

	f := framework.Global

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	noUsername := makeHost(t, ctx, "missing-bmc-username",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-no-user",
			},
		})
	waitForErrorStatus(t, noUsername)

	secret := &corev1.Secret{}
	secretName := types.NamespacedName{
		Namespace: namespace,
		Name:      "bmc-creds-no-user",
	}
	err = f.Client.Get(goctx.TODO(), secretName, secret)
	if err != nil {
		t.Fatal(err)
	}
	secret.Data["username"] = []byte(base64.StdEncoding.EncodeToString([]byte("username")))
	err = f.Client.Update(goctx.TODO(), secret)
	if err != nil {
		t.Fatal(err)
	}
	waitForOfflineStatus(t, noUsername)
}

func TestSetOffline(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "toggle-offline",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
			Online: true,
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState before toggle: %s", state)
		if state == metalkubev1alpha1.OperationalStatusOnline {
			return true, nil
		}
		return false, nil
	})

	refreshHost(host)
	host.Spec.Online = false
	f := framework.Global
	err := f.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState after toggle: %s", state)
		if state == metalkubev1alpha1.OperationalStatusOffline {
			return true, nil
		}
		return false, nil
	})

}

func TestSetOnline(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "toggle-online",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
			Online: false,
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState before toggle: %s", state)
		if state == metalkubev1alpha1.OperationalStatusOffline {
			return true, nil
		}
		return false, nil
	})

	refreshHost(host)
	host.Spec.Online = true
	f := framework.Global
	err := f.Client.Update(goctx.TODO(), host)
	if err != nil {
		t.Fatal(err)
	}

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		state := host.Labels[metalkubev1alpha1.OperationalStatusLabel]
		t.Logf("OperationalState after toggle: %s", state)
		if state == metalkubev1alpha1.OperationalStatusOnline {
			return true, nil
		}
		return false, nil
	})

}

func TestSetHardwareProfileLabel(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	host := makeHost(t, ctx, "hardware-profile",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("labels: %v", host.ObjectMeta.Labels)
		if host.ObjectMeta.Labels[metalkubev1alpha1.HardwareProfileLabel] != "" {
			return true, nil
		}
		return false, nil
	})
}

func TestManageHardwareDetails(t *testing.T) {
	ctx := setup(t)
	defer ctx.Cleanup()

	f := framework.Global

	host := makeHost(t, ctx, "hardware-profile",
		&metalkubev1alpha1.BareMetalHostSpec{
			BMC: metalkubev1alpha1.BMCDetails{
				IP:              "192.168.100.100",
				CredentialsName: "bmc-creds-valid",
			},
		})

	// Details should be filled in when the host is created...
	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("details: %v", host.Status.HardwareDetails)
		if host.Status.HardwareDetails != nil {
			return true, nil
		}
		return false, nil
	})

	if err := f.Client.Delete(goctx.TODO(), host); err != nil {
		t.Fatal(err)
	}

	// and removed when the host is deleted.
	waitForHostStateChange(t, host, func(host *metalkubev1alpha1.BareMetalHost) (done bool, err error) {
		t.Logf("details: %v", host.Status.HardwareDetails)
		if host.Status.HardwareDetails == nil {
			return true, nil
		}
		return false, nil
	})
}
