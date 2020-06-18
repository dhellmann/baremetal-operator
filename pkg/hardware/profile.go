package hardware

import (
	"fmt"

	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
)

const (
	// DefaultProfileName is the default hardware profile to use when
	// no other profile matches.
	DefaultProfileName string = "unknown"
)

// Profile holds the settings for a class of hardware.
type Profile struct {
	// Name holds the profile name
	Name string

	// RootDeviceHints holds the suggestions for placing the storage
	// for the root filesystem.
	RootDeviceHints metal3shared.RootDeviceHints
}

var profiles = make(map[string]Profile)

func init() {
	profiles[DefaultProfileName] = Profile{
		Name: DefaultProfileName,
		RootDeviceHints: metal3shared.RootDeviceHints{
			DeviceName: "/dev/sda",
		},
	}

	profiles["libvirt"] = Profile{
		Name: "libvirt",
		RootDeviceHints: metal3shared.RootDeviceHints{
			DeviceName: "/dev/vda",
		},
	}

	profiles["dell"] = Profile{
		Name: "dell",
		RootDeviceHints: metal3shared.RootDeviceHints{
			HCTL: "0:0:0:0",
		},
	}

	profiles["dell-raid"] = Profile{
		Name: "dell-raid",
		RootDeviceHints: metal3shared.RootDeviceHints{
			HCTL: "0:2:0:0",
		},
	}

	profiles["openstack"] = Profile{
		Name: "openstack",
		RootDeviceHints: metal3shared.RootDeviceHints{
			DeviceName: "/dev/vdb",
		},
	}
}

// GetProfile returns the named profile
func GetProfile(name string) (Profile, error) {
	profile, ok := profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("No hardware profile named %q", name)
	}
	return profile, nil
}
