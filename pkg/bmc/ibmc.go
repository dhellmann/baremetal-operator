package bmc

import (
	"net/url"
	"strings"

	metal3v1alpha1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"
)

func init() {
	registerFactory("ibmc", newIbmcAccessDetails, []string{"http", "https"})
}

func newIbmcAccessDetails(parsedURL *url.URL, disableCertificateVerification bool) (AccessDetails, error) {
	return &ibmcAccessDetails{
		bmcType:                        parsedURL.Scheme,
		host:                           parsedURL.Host,
		path:                           parsedURL.Path,
		disableCertificateVerification: disableCertificateVerification,
	}, nil
}

type ibmcAccessDetails struct {
	bmcType                        string
	host                           string
	path                           string
	disableCertificateVerification bool
}

func (a *ibmcAccessDetails) Type() string {
	return a.bmcType
}

// NeedsMAC returns true when the host is going to need a separate
// port created rather than having it discovered.
func (a *ibmcAccessDetails) NeedsMAC() bool {
	// For the inspection to work, we need a MAC address
	// https://github.com/metal3-io/baremetal-operator/pull/284#discussion_r317579040
	return true
}

func (a *ibmcAccessDetails) Driver() string {
	return "ibmc"
}

func (a *ibmcAccessDetails) DisableCertificateVerification() bool {
	return a.disableCertificateVerification
}

const ibmcDefaultScheme = "https"

// DriverInfo returns a data structure to pass as the DriverInfo
// parameter when creating a node in Ironic. The structure is
// pre-populated with the access information, and the caller is
// expected to add any other information that might be needed (such as
// the kernel and ramdisk locations).
func (a *ibmcAccessDetails) DriverInfo(bmcCreds Credentials) map[string]interface{} {

	ibmcAddress := []string{}
	schemes := strings.Split(a.bmcType, "+")
	if len(schemes) > 1 {
		ibmcAddress = append(ibmcAddress, schemes[1])
	} else {
		ibmcAddress = append(ibmcAddress, ibmcDefaultScheme)
	}
	ibmcAddress = append(ibmcAddress, "://")
	ibmcAddress = append(ibmcAddress, a.host)
	ibmcAddress = append(ibmcAddress, a.path)

	result := map[string]interface{}{
		"ibmc_username": bmcCreds.Username,
		"ibmc_password": bmcCreds.Password,
		"ibmc_address":  strings.Join(ibmcAddress, ""),
	}

	if a.disableCertificateVerification {
		result["ibmc_verify_ca"] = false
	}

	return result
}

// NodeProperties returns a data structure to return details of
// the host, including the boot mode. This will be used later to
// instruct ironic to use specific boot mode
func (a *ibmcAccessDetails) NodeProperties() map[string]interface{} {
	result := map[string]interface{}{
		"boot_mode": metal3v1alpha1.UEFI,
	}
	return result
}

func (a *ibmcAccessDetails) BootInterface() string {
	return "pxe"
}

func (a *ibmcAccessDetails) ManagementInterface() string {
	return "ibmc"
}

func (a *ibmcAccessDetails) PowerInterface() string {
	return "ibmc"
}

func (a *ibmcAccessDetails) RAIDInterface() string {
	return ""
}

func (a *ibmcAccessDetails) VendorInterface() string {
	return ""
}
