// +build !ignore_autogenerated

/*
Copyright 2018 The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceWrapper) DeepCopyInto(out *ServiceWrapper) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceWrapper.
func (in *ServiceWrapper) DeepCopy() *ServiceWrapper {
	if in == nil {
		return nil
	}
	out := new(ServiceWrapper)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishService) DeepCopyInto(out *VarnishService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishService.
func (in *VarnishService) DeepCopy() *VarnishService {
	if in == nil {
		return nil
	}
	out := new(VarnishService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VarnishService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishServiceList) DeepCopyInto(out *VarnishServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VarnishService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishServiceList.
func (in *VarnishServiceList) DeepCopy() *VarnishServiceList {
	if in == nil {
		return nil
	}
	out := new(VarnishServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VarnishServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishServiceSpec) DeepCopyInto(out *VarnishServiceSpec) {
	*out = *in
	in.Service.DeepCopyInto(&out.Service)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishServiceSpec.
func (in *VarnishServiceSpec) DeepCopy() *VarnishServiceSpec {
	if in == nil {
		return nil
	}
	out := new(VarnishServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VarnishServiceStatus) DeepCopyInto(out *VarnishServiceStatus) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VarnishServiceStatus.
func (in *VarnishServiceStatus) DeepCopy() *VarnishServiceStatus {
	if in == nil {
		return nil
	}
	out := new(VarnishServiceStatus)
	in.DeepCopyInto(out)
	return out
}
