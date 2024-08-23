/*
Copyright The Kubernetes Authors.

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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
	apisv1alpha1 "sigs.k8s.io/network-policy-api/apis/v1alpha1"
)

// BaselineAdminNetworkPolicyEgressPeerApplyConfiguration represents an declarative configuration of the BaselineAdminNetworkPolicyEgressPeer type for use
// with apply.
type BaselineAdminNetworkPolicyEgressPeerApplyConfiguration struct {
	Namespaces *v1.LabelSelectorApplyConfiguration `json:"namespaces,omitempty"`
	Pods       *NamespacedPodApplyConfiguration    `json:"pods,omitempty"`
	Nodes      *v1.LabelSelectorApplyConfiguration `json:"nodes,omitempty"`
	Networks   []apisv1alpha1.CIDR                 `json:"networks,omitempty"`
}

// BaselineAdminNetworkPolicyEgressPeerApplyConfiguration constructs an declarative configuration of the BaselineAdminNetworkPolicyEgressPeer type for use with
// apply.
func BaselineAdminNetworkPolicyEgressPeer() *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration {
	return &BaselineAdminNetworkPolicyEgressPeerApplyConfiguration{}
}

// WithNamespaces sets the Namespaces field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespaces field is set to the value of the last call.
func (b *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration) WithNamespaces(value *v1.LabelSelectorApplyConfiguration) *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration {
	b.Namespaces = value
	return b
}

// WithPods sets the Pods field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Pods field is set to the value of the last call.
func (b *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration) WithPods(value *NamespacedPodApplyConfiguration) *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration {
	b.Pods = value
	return b
}

// WithNodes sets the Nodes field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Nodes field is set to the value of the last call.
func (b *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration) WithNodes(value *v1.LabelSelectorApplyConfiguration) *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration {
	b.Nodes = value
	return b
}

// WithNetworks adds the given value to the Networks field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Networks field.
func (b *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration) WithNetworks(values ...apisv1alpha1.CIDR) *BaselineAdminNetworkPolicyEgressPeerApplyConfiguration {
	for i := range values {
		b.Networks = append(b.Networks, values[i])
	}
	return b
}
