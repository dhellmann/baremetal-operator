package ironic

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/gophercloud/gophercloud/openstack/baremetal/v1/nodes"

	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
	metal3 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha2"
	"github.com/metal3-io/baremetal-operator/pkg/bmc"
	"k8s.io/utils/pointer"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestGetUpdateOptsForNodeWithRootHints(t *testing.T) {
	rotational := true
	host := &metal3.BareMetalHost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myhost",
			Namespace: "myns",
			UID:       "27720611-e5d1-45d3-ba3a-222dcfaa4ca2",
		},
		Spec: metal3.BareMetalHostSpec{
			Image: &metal3shared.Image{
				URL: "not-empty",
			},
			Online:          true,
			HardwareProfile: "libvirt",
			RootDeviceHints: &metal3shared.RootDeviceHints{
				DeviceName:         "userd_devicename",
				HCTL:               "1:2:3:4",
				Model:              "userd_model",
				Vendor:             "userd_vendor",
				SerialNumber:       "userd_serial",
				MinSizeGigabytes:   40,
				WWN:                "userd_wwn",
				WWNWithExtension:   "userd_with_extension",
				WWNVendorExtension: "userd_vendor_extension",
				Rotational:         &rotational,
			},
		},
		Status: metal3.BareMetalHostStatus{
			Provisioning: metal3shared.ProvisionStatus{
				ID: "provisioning-id",
				// Place the hints in the status field to pretend the
				// controller has already reconciled partially.
				RootDeviceHints: &metal3shared.RootDeviceHints{
					DeviceName:         "userd_devicename",
					HCTL:               "1:2:3:4",
					Model:              "userd_model",
					Vendor:             "userd_vendor",
					SerialNumber:       "userd_serial",
					MinSizeGigabytes:   40,
					WWN:                "userd_wwn",
					WWNWithExtension:   "userd_with_extension",
					WWNVendorExtension: "userd_vendor_extension",
					Rotational:         &rotational,
				},
			},
			HardwareProfile: "libvirt",
		},
	}

	eventPublisher := func(reason, message string) {}

	prov, err := newProvisioner(host, bmc.Credentials{}, eventPublisher)
	ironicNode := &nodes.Node{}

	patches, err := prov.getUpdateOptsForNode(ironicNode)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("patches: %v", patches)

	expected := []struct {
		Path  string            // the node property path
		Map   map[string]string // Expected roothdevicehint map
		Value interface{}       // the value being passed to ironic (or value associated with the key)
	}{
		{
			Path:  "/properties/root_device",
			Value: "userdefined_devicename",
			Map: map[string]string{
				"name":                 "s== userd_devicename",
				"hctl":                 "s== 1:2:3:4",
				"model":                "<in> userd_model",
				"vendor":               "<in> userd_vendor",
				"serial":               "s== userd_serial",
				"size":                 ">= 40",
				"wwn":                  "s== userd_wwn",
				"wwn_with_extension":   "s== userd_with_extension",
				"wwn_vendor_extension": "s== userd_vendor_extension",
				"rotational":           "true",
			},
		},
	}

	for _, e := range expected {
		t.Run(e.Path, func(t *testing.T) {
			t.Logf("expected: %v", e)
			var update nodes.UpdateOperation
			for _, patch := range patches {
				update = patch.(nodes.UpdateOperation)
				if update.Path == e.Path {
					break
				}
			}
			if update.Path != e.Path {
				t.Errorf("did not find %q in updates", e.Path)
				return
			}
			t.Logf("update: %v", update)
			if e.Map != nil {
				assert.Equal(t, e.Map, update.Value, fmt.Sprintf("%s does not match", e.Path))
			} else {
				assert.Equal(t, e.Value, update.Value, fmt.Sprintf("%s does not match", e.Path))
			}
		})
	}
}

func TestGetUpdateOptsForNodeVirtual(t *testing.T) {
	host := &metal3.BareMetalHost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myhost",
			Namespace: "myns",
			UID:       "27720611-e5d1-45d3-ba3a-222dcfaa4ca2",
		},
		Spec: metal3.BareMetalHostSpec{
			Image: &metal3shared.Image{
				URL:          "not-empty",
				Checksum:     "checksum",
				ChecksumType: metal3shared.MD5,
				DiskFormat:   pointer.StringPtr("raw"),
			},
			Online: true,
		},
		Status: metal3.BareMetalHostStatus{
			HardwareProfile: "libvirt",
			Provisioning: metal3shared.ProvisionStatus{
				ID: "provisioning-id",
			},
		},
	}

	eventPublisher := func(reason, message string) {}

	prov, err := newProvisioner(host, bmc.Credentials{}, eventPublisher)
	ironicNode := &nodes.Node{}

	patches, err := prov.getUpdateOptsForNode(ironicNode)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("patches: %v", patches)

	expected := []struct {
		Path  string      // the node property path
		Key   string      // if value is a map, the key we care about
		Value interface{} // the value being passed to ironic (or value associated with the key)
	}{
		{
			Path:  "/instance_info/image_source",
			Value: "not-empty",
		},
		{
			Path:  "/instance_info/image_os_hash_algo",
			Value: "md5",
		},
		{
			Path:  "/instance_info/image_os_hash_value",
			Value: "checksum",
		},
		{
			Path:  "/instance_info/image_disk_format",
			Value: "raw",
		},
		{
			Path:  "/instance_uuid",
			Value: "27720611-e5d1-45d3-ba3a-222dcfaa4ca2",
		},
		{
			Path:  "/instance_info/root_gb",
			Value: 10,
		},
		{
			Path:  "/properties/cpu_arch",
			Value: "x86_64",
		},
		{
			Path:  "/properties/local_gb",
			Value: 50,
		},
	}

	for _, e := range expected {
		t.Run(e.Path, func(t *testing.T) {
			t.Logf("expected: %v", e)
			var update nodes.UpdateOperation
			for _, patch := range patches {
				update = patch.(nodes.UpdateOperation)
				if update.Path == e.Path {
					break
				}
			}
			if update.Path != e.Path {
				t.Errorf("did not find %q in updates", e.Path)
				return
			}
			t.Logf("update: %v", update)
			assert.Equal(t, e.Value, update.Value, fmt.Sprintf("%s does not match", e.Path))
		})
	}
}

func TestGetUpdateOptsForNodeDell(t *testing.T) {
	host := &metal3.BareMetalHost{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myhost",
			Namespace: "myns",
			UID:       "27720611-e5d1-45d3-ba3a-222dcfaa4ca2",
		},
		Spec: metal3.BareMetalHostSpec{
			Image: &metal3shared.Image{
				URL:          "not-empty",
				Checksum:     "checksum",
				ChecksumType: metal3shared.MD5,
				//DiskFormat not given to verify it is not added in instance_info
			},
			Online: true,
		},
		Status: metal3.BareMetalHostStatus{
			HardwareProfile: "dell",
			Provisioning: metal3shared.ProvisionStatus{
				ID: "provisioning-id",
			},
		},
	}

	eventPublisher := func(reason, message string) {}

	prov, err := newProvisioner(host, bmc.Credentials{}, eventPublisher)
	ironicNode := &nodes.Node{}

	patches, err := prov.getUpdateOptsForNode(ironicNode)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("patches: %v", patches)

	expected := []struct {
		Path  string      // the node property path
		Key   string      // if value is a map, the key we care about
		Value interface{} // the value being passed to ironic (or value associated with the key)
	}{
		{
			Path:  "/instance_info/image_source",
			Value: "not-empty",
		},
		{
			Path:  "/instance_info/image_os_hash_algo",
			Value: "md5",
		},
		{
			Path:  "/instance_info/image_os_hash_value",
			Value: "checksum",
		},
		{
			Path:  "/instance_uuid",
			Value: "27720611-e5d1-45d3-ba3a-222dcfaa4ca2",
		},
		{
			Path:  "/instance_info/root_gb",
			Value: 10,
		},
		{
			Path:  "/properties/cpu_arch",
			Value: "x86_64",
		},
		{
			Path:  "/properties/local_gb",
			Value: 50,
		},
	}

	for _, e := range expected {
		t.Run(e.Path, func(t *testing.T) {
			t.Logf("expected: %v", e)
			var update nodes.UpdateOperation
			for _, patch := range patches {
				update = patch.(nodes.UpdateOperation)
				if update.Path == e.Path {
					break
				}
			}
			if update.Path != e.Path {
				t.Errorf("did not find %q in updates", e.Path)
				return
			}
			t.Logf("update: %v", update)
			assert.Equal(t, e.Value, update.Value, fmt.Sprintf("%s does not match", e.Path))
		})
	}
}
