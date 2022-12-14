//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 Bernhard Aichinger.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActualState) DeepCopyInto(out *ActualState) {
	*out = *in
	out.Version = in.Version
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActualState.
func (in *ActualState) DeepCopy() *ActualState {
	if in == nil {
		return nil
	}
	out := new(ActualState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DesiredState) DeepCopyInto(out *DesiredState) {
	*out = *in
	out.Version = in.Version
	in.UpdatedTimestamp.DeepCopyInto(&out.UpdatedTimestamp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DesiredState.
func (in *DesiredState) DeepCopy() *DesiredState {
	if in == nil {
		return nil
	}
	out := new(DesiredState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Paper) DeepCopyInto(out *Paper) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Paper.
func (in *Paper) DeepCopy() *Paper {
	if in == nil {
		return nil
	}
	out := new(Paper)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Paper) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PaperList) DeepCopyInto(out *PaperList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Paper, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PaperList.
func (in *PaperList) DeepCopy() *PaperList {
	if in == nil {
		return nil
	}
	out := new(PaperList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PaperList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PaperSpec) DeepCopyInto(out *PaperSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PaperSpec.
func (in *PaperSpec) DeepCopy() *PaperSpec {
	if in == nil {
		return nil
	}
	out := new(PaperSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PaperStatus) DeepCopyInto(out *PaperStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DesiredState != nil {
		in, out := &in.DesiredState, &out.DesiredState
		*out = new(DesiredState)
		(*in).DeepCopyInto(*out)
	}
	if in.ActualState != nil {
		in, out := &in.ActualState, &out.ActualState
		*out = new(ActualState)
		**out = **in
	}
	if in.UpdatedTimestamp != nil {
		in, out := &in.UpdatedTimestamp, &out.UpdatedTimestamp
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PaperStatus.
func (in *PaperStatus) DeepCopy() *PaperStatus {
	if in == nil {
		return nil
	}
	out := new(PaperStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Version) DeepCopyInto(out *Version) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Version.
func (in *Version) DeepCopy() *Version {
	if in == nil {
		return nil
	}
	out := new(Version)
	in.DeepCopyInto(out)
	return out
}
