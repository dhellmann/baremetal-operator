// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha2

import (
	shared "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/shared"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BareMetalHost) DeepCopyInto(out *BareMetalHost) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
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
	return
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
		*out = new(shared.RootDeviceHints)
		(*in).DeepCopyInto(*out)
	}
	if in.ConsumerRef != nil {
		in, out := &in.ConsumerRef, &out.ConsumerRef
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(shared.Image)
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
	return
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
	return
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
	return
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
func (in *HardwareDetails) DeepCopyInto(out *HardwareDetails) {
	*out = *in
	out.SystemVendor = in.SystemVendor
	out.Firmware = in.Firmware
	if in.NIC != nil {
		in, out := &in.NIC, &out.NIC
		*out = make([]shared.NIC, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = make([]shared.Storage, len(*in))
		copy(*out, *in)
	}
	in.CPU.DeepCopyInto(&out.CPU)
	return
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
func (in *ProvisionStatus) DeepCopyInto(out *ProvisionStatus) {
	*out = *in
	in.Image.DeepCopyInto(&out.Image)
	if in.RootDeviceHints != nil {
		in, out := &in.RootDeviceHints, &out.RootDeviceHints
		*out = new(shared.RootDeviceHints)
		(*in).DeepCopyInto(*out)
	}
	return
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
