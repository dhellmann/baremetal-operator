package hardwaredetails

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/baremetalintrospection/v1/introspection"
	metal3shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
	metal3 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha2"
)

// GetHardwareDetails converts Ironic introspection data into BareMetalHost HardwareDetails.
func GetHardwareDetails(data *introspection.Data) *metal3.HardwareDetails {
	details := new(metal3.HardwareDetails)
	details.Firmware = getFirmwareDetails(data.Extra.Firmware)
	details.SystemVendor = getSystemVendorDetails(data.Inventory.SystemVendor)
	details.RAMMebibytes = data.MemoryMB
	details.NIC = getNICDetails(data.Inventory.Interfaces, data.AllInterfaces, data.Extra.Network)
	details.Storage = getStorageDetails(data.Inventory.Disks)
	details.CPU = getCPUDetails(&data.Inventory.CPU)
	details.Hostname = data.Inventory.Hostname
	return details
}

func getVLANs(intf introspection.BaseInterfaceType) (vlans []metal3shared.VLAN, vlanid metal3shared.VLANID) {
	if intf.LLDPProcessed == nil {
		return
	}
	if spvs, ok := intf.LLDPProcessed["switch_port_vlans"]; ok {
		if data, ok := spvs.([]map[string]interface{}); ok {
			vlans = make([]metal3shared.VLAN, len(data))
			for i, vlan := range data {
				vid, _ := vlan["id"].(int)
				name, _ := vlan["name"].(string)
				vlans[i] = metal3shared.VLAN{
					ID:   metal3shared.VLANID(vid),
					Name: name,
				}
			}
		}
	}
	if vid, ok := intf.LLDPProcessed["switch_port_untagged_vlan_id"].(int); ok {
		vlanid = metal3shared.VLANID(vid)
	}
	return
}

func getNICSpeedGbps(intfExtradata introspection.ExtraHardwareData) (speedGbps int) {
	if speed, ok := intfExtradata["speed"].(string); ok {
		if strings.HasSuffix(speed, "Gbps") {
			fmt.Sscanf(speed, "%d", &speedGbps)
		}
	}
	return
}

func getNICDetails(ifdata []introspection.InterfaceType,
	basedata map[string]introspection.BaseInterfaceType,
	extradata introspection.ExtraHardwareDataSection) []metal3shared.NIC {
	nics := make([]metal3shared.NIC, len(ifdata))
	for i, intf := range ifdata {
		baseIntf := basedata[intf.Name]
		vlans, vlanid := getVLANs(baseIntf)
		ip := intf.IPV4Address
		if ip == "" {
			ip = intf.IPV6Address
		}
		nics[i] = metal3shared.NIC{
			Name: intf.Name,
			Model: strings.TrimLeft(fmt.Sprintf("%s %s",
				intf.Vendor, intf.Product), " "),
			MAC:       intf.MACAddress,
			IP:        ip,
			VLANs:     vlans,
			VLANID:    vlanid,
			SpeedGbps: getNICSpeedGbps(extradata[intf.Name]),
			PXE:       baseIntf.PXE,
		}
	}
	return nics
}

func getStorageDetails(diskdata []introspection.RootDiskType) []metal3shared.Storage {
	storage := make([]metal3shared.Storage, len(diskdata))
	for i, disk := range diskdata {
		storage[i] = metal3shared.Storage{
			Name:               disk.Name,
			Rotational:         disk.Rotational,
			SizeBytes:          metal3shared.Capacity(disk.Size),
			Vendor:             disk.Vendor,
			Model:              disk.Model,
			SerialNumber:       disk.Serial,
			WWN:                disk.Wwn,
			WWNVendorExtension: disk.WwnVendorExtension,
			WWNWithExtension:   disk.WwnWithExtension,
			HCTL:               disk.Hctl,
		}
	}
	return storage
}

func getSystemVendorDetails(vendor introspection.SystemVendorType) metal3shared.HardwareSystemVendor {
	return metal3shared.HardwareSystemVendor{
		Manufacturer: vendor.Manufacturer,
		ProductName:  vendor.ProductName,
		SerialNumber: vendor.SerialNumber,
	}
}

func getCPUDetails(cpudata *introspection.CPUType) metal3.CPU {
	var freq float64
	fmt.Sscanf(cpudata.Frequency, "%f", &freq)
	sort.Strings(cpudata.Flags)
	cpu := metal3.CPU{
		Arch:           cpudata.Architecture,
		Model:          cpudata.ModelName,
		ClockMegahertz: metal3.ClockSpeed(freq) * metal3.MegaHertz,
		Count:          cpudata.Count,
		Flags:          cpudata.Flags,
	}

	return cpu
}

func getFirmwareDetails(firmwaredata introspection.ExtraHardwareDataSection) metal3shared.Firmware {

	// handle bios optionally
	var bios metal3shared.BIOS

	if biosdata, ok := firmwaredata["bios"]; ok {
		// we do not know if all fields will be supplied
		// as this is not a structured response
		// so we must handle each field conditionally
		bios.Vendor, _ = biosdata["vendor"].(string)
		bios.Version, _ = biosdata["version"].(string)
		bios.Date, _ = biosdata["date"].(string)
	}

	return metal3shared.Firmware{
		BIOS: bios,
	}

}
