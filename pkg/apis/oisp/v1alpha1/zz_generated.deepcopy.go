// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OispDevicesManager) DeepCopyInto(out *OispDevicesManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OispDevicesManager.
func (in *OispDevicesManager) DeepCopy() *OispDevicesManager {
	if in == nil {
		return nil
	}
	out := new(OispDevicesManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OispDevicesManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OispDevicesManagerList) DeepCopyInto(out *OispDevicesManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OispDevicesManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OispDevicesManagerList.
func (in *OispDevicesManagerList) DeepCopy() *OispDevicesManagerList {
	if in == nil {
		return nil
	}
	out := new(OispDevicesManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OispDevicesManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OispDevicesManagerSpec) DeepCopyInto(out *OispDevicesManagerSpec) {
	*out = *in
	in.PodTemplateSpec.DeepCopyInto(&out.PodTemplateSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OispDevicesManagerSpec.
func (in *OispDevicesManagerSpec) DeepCopy() *OispDevicesManagerSpec {
	if in == nil {
		return nil
	}
	out := new(OispDevicesManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OispDevicesManagerStatus) DeepCopyInto(out *OispDevicesManagerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OispDevicesManagerStatus.
func (in *OispDevicesManagerStatus) DeepCopy() *OispDevicesManagerStatus {
	if in == nil {
		return nil
	}
	out := new(OispDevicesManagerStatus)
	in.DeepCopyInto(out)
	return out
}
