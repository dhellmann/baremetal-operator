// +build !ignore_autogenerated

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BIOS) DeepCopyInto(out *BIOS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BIOS.
func (in *BIOS) DeepCopy() *BIOS {
	if in == nil {
		return nil
	}
	out := new(BIOS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BMCDetails) DeepCopyInto(out *BMCDetails) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BMCDetails.
func (in *BMCDetails) DeepCopy() *BMCDetails {
	if in == nil {
		return nil
	}
	out := new(BMCDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BareMetalHost) DeepCopyInto(out *BareMetalHost) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BareMetalHost.
func (in *BareMetalHost) DeepCopy() *BareMetalHost {
	if in == nil {
		return nil
	}
	out := new(BareMetalHost)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BareMetalHost) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BareMetalHostList) DeepCopyInto(out *BareMetalHostList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]BareMetalHost, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BareMetalHostList.
func (in *BareMetalHostList) DeepCopy() *BareMetalHostList {
	if in == nil {
		return nil
	}
	out := new(BareMetalHostList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BareMetalHostList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BareMetalHostSpec) DeepCopyInto(out *BareMetalHostSpec) {
	*out = *in
	if in.Taints != nil {
		in, out := &in.Taints, &out.Taints
		*out = make([]v1.Taint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.BMC = in.BMC
	if in.RootDeviceHints != nil {
		in, out := &in.RootDeviceHints, &out.RootDeviceHints
		*out = new(RootDeviceHints)
		(*in).DeepCopyInto(*out)
	}
	if in.ConsumerRef != nil {
		in, out := &in.ConsumerRef, &out.ConsumerRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(Image)
		(*in).DeepCopyInto(*out)
	}
	if in.UserData != nil {
		in, out := &in.UserData, &out.UserData
		*out = new(v1.SecretReference)
		**out = **in
	}
	if in.NetworkData != nil {
		in, out := &in.NetworkData, &out.NetworkData
		*out = new(v1.SecretReference)
		**out = **in
	}
	if in.MetaData != nil {
		in, out := &in.MetaData, &out.MetaData
		*out = new(v1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BareMetalHostSpec.
func (in *BareMetalHostSpec) DeepCopy() *BareMetalHostSpec {
	if in == nil {
		return nil
	}
	out := new(BareMetalHostSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BareMetalHostStatus) DeepCopyInto(out *BareMetalHostStatus) {
	*out = *in
	if in.LastUpdated != nil {
		in, out := &in.LastUpdated, &out.LastUpdated
		*out = (*in).DeepCopy()
	}
	if in.HardwareDetails != nil {
		in, out := &in.HardwareDetails, &out.HardwareDetails
		*out = new(HardwareDetails)
		(*in).DeepCopyInto(*out)
	}
	in.Provisioning.DeepCopyInto(&out.Provisioning)
	in.GoodCredentials.DeepCopyInto(&out.GoodCredentials)
	in.TriedCredentials.DeepCopyInto(&out.TriedCredentials)
	in.OperationHistory.DeepCopyInto(&out.OperationHistory)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BareMetalHostStatus.
func (in *BareMetalHostStatus) DeepCopy() *BareMetalHostStatus {
	if in == nil {
		return nil
	}
	out := new(BareMetalHostStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CPU) DeepCopyInto(out *CPU) {
	*out = *in
	if in.Flags != nil {
		in, out := &in.Flags, &out.Flags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CPU.
func (in *CPU) DeepCopy() *CPU {
	if in == nil {
		return nil
	}
	out := new(CPU)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CredentialsStatus) DeepCopyInto(out *CredentialsStatus) {
	*out = *in
	if in.Reference != nil {
		in, out := &in.Reference, &out.Reference
		*out = new(v1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CredentialsStatus.
func (in *CredentialsStatus) DeepCopy() *CredentialsStatus {
	if in == nil {
		return nil
	}
	out := new(CredentialsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Firmware) DeepCopyInto(out *Firmware) {
	*out = *in
	out.BIOS = in.BIOS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Firmware.
func (in *Firmware) DeepCopy() *Firmware {
	if in == nil {
		return nil
	}
	out := new(Firmware)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HardwareDetails) DeepCopyInto(out *HardwareDetails) {
	*out = *in
	out.SystemVendor = in.SystemVendor
	out.Firmware = in.Firmware
	if in.NIC != nil {
		in, out := &in.NIC, &out.NIC
		*out = make([]NIC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = make([]Storage, len(*in))
		copy(*out, *in)
	}
	in.CPU.DeepCopyInto(&out.CPU)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HardwareDetails.
func (in *HardwareDetails) DeepCopy() *HardwareDetails {
	if in == nil {
		return nil
	}
	out := new(HardwareDetails)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HardwareSystemVendor) DeepCopyInto(out *HardwareSystemVendor) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HardwareSystemVendor.
func (in *HardwareSystemVendor) DeepCopy() *HardwareSystemVendor {
	if in == nil {
		return nil
	}
	out := new(HardwareSystemVendor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Image) DeepCopyInto(out *Image) {
	*out = *in
	if in.DiskFormat != nil {
		in, out := &in.DiskFormat, &out.DiskFormat
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Image.
func (in *Image) DeepCopy() *Image {
	if in == nil {
		return nil
	}
	out := new(Image)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NIC) DeepCopyInto(out *NIC) {
	*out = *in
	if in.VLANs != nil {
		in, out := &in.VLANs, &out.VLANs
		*out = make([]VLAN, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NIC.
func (in *NIC) DeepCopy() *NIC {
	if in == nil {
		return nil
	}
	out := new(NIC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperationHistory) DeepCopyInto(out *OperationHistory) {
	*out = *in
	in.Register.DeepCopyInto(&out.Register)
	in.Inspect.DeepCopyInto(&out.Inspect)
	in.Provision.DeepCopyInto(&out.Provision)
	in.Deprovision.DeepCopyInto(&out.Deprovision)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperationHistory.
func (in *OperationHistory) DeepCopy() *OperationHistory {
	if in == nil {
		return nil
	}
	out := new(OperationHistory)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperationMetric) DeepCopyInto(out *OperationMetric) {
	*out = *in
	in.Start.DeepCopyInto(&out.Start)
	in.End.DeepCopyInto(&out.End)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperationMetric.
func (in *OperationMetric) DeepCopy() *OperationMetric {
	if in == nil {
		return nil
	}
	out := new(OperationMetric)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProvisionStatus) DeepCopyInto(out *ProvisionStatus) {
	*out = *in
	in.Image.DeepCopyInto(&out.Image)
	if in.RootDeviceHints != nil {
		in, out := &in.RootDeviceHints, &out.RootDeviceHints
		*out = new(RootDeviceHints)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProvisionStatus.
func (in *ProvisionStatus) DeepCopy() *ProvisionStatus {
	if in == nil {
		return nil
	}
	out := new(ProvisionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RootDeviceHints) DeepCopyInto(out *RootDeviceHints) {
	*out = *in
	if in.Rotational != nil {
		in, out := &in.Rotational, &out.Rotational
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RootDeviceHints.
func (in *RootDeviceHints) DeepCopy() *RootDeviceHints {
	if in == nil {
		return nil
	}
	out := new(RootDeviceHints)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Storage) DeepCopyInto(out *Storage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Storage.
func (in *Storage) DeepCopy() *Storage {
	if in == nil {
		return nil
	}
	out := new(Storage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VLAN) DeepCopyInto(out *VLAN) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VLAN.
func (in *VLAN) DeepCopy() *VLAN {
	if in == nil {
		return nil
	}
	out := new(VLAN)
	in.DeepCopyInto(out)
	return out
}
