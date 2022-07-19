/*
Copyright 2014 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/attachinterfaces"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	v1 "k8s.io/api/core/v1"
)

func buildNodeAddressesTestConfig() (*servers.Server, *[]attachinterfaces.Interface) {
	server := servers.Server{
		Status:     "ACTIVE",
		HostID:     "29d3c8c896a45aa4c34e52247875d7fefc3d94bbcc9f622b5d204362",
		AccessIPv4: "50.56.176.99",
		AccessIPv6: "2001:4800:790e:510:be76:4eff:fe04:82a8",
		Addresses: map[string]interface{}{
			"private": []interface{}{
				map[string]interface{}{
					"OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:7c:1b:2b",
					"version":                 float64(4),
					"addr":                    "10.0.0.32",
					"OS-EXT-IPS:type":         "fixed",
				},
				map[string]interface{}{
					"version":         float64(4),
					"addr":            "50.56.176.36",
					"OS-EXT-IPS:type": "floating",
				},
				map[string]interface{}{
					"version": float64(4),
					"addr":    "10.0.0.31",
					// No OS-EXT-IPS:type
				},
			},
			"public": []interface{}{
				map[string]interface{}{
					"version": float64(4),
					"addr":    "50.56.176.35",
				},
				map[string]interface{}{
					"version": float64(6),
					"addr":    "2001:4800:780e:510:be76:4eff:fe04:84a8",
				},
			},
		},
		Metadata: map[string]string{
			"name":       "a1-yinvcez57-0-bvynoyawrhcg-kube-minion-fg5i4jwcc2yy",
			TypeHostName: "a1-yinvcez57-0-bvynoyawrhcg-kube-minion-fg5i4jwcc2yy.novalocal",
		},
	}

	interfaces := []attachinterfaces.Interface{
		{
			PortState: "ACTIVE",
			FixedIPs: []attachinterfaces.FixedIP{
				{
					IPAddress: "10.0.0.32",
				},
				{
					IPAddress: "10.0.0.31",
				},
			},
		},
	}

	return &server, &interfaces
}

func TestNodeAddressesPreferredOrderIPv4BeforeIPv6(t *testing.T) {
	srv, interfaces := buildNodeAddressesTestConfig()

	networkingOpts := NetworkingOpts{
		PublicNetworkName:     []string{"public"},
		PreferredAddressOrder: fmt.Sprintf("%s, %s, %s, %s", IPv4InternalOrderID, IPv4ExternalOrderID, IPv6InternalOrderID, IPv6ExternalOrderID),
	}

	addrs, err := nodeAddresses(srv, *interfaces, networkingOpts)
	if err != nil {
		t.Fatalf("nodeAddresses returned error: %v", err)
	}

	t.Logf("addresses are %v", addrs)

	want := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: "10.0.0.32"},
		{Type: v1.NodeInternalIP, Address: "10.0.0.31"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.99"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.36"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.35"},
		{Type: v1.NodeExternalIP, Address: "2001:4800:790e:510:be76:4eff:fe04:82a8"},
		{Type: v1.NodeExternalIP, Address: "2001:4800:780e:510:be76:4eff:fe04:84a8"},
		{Type: v1.NodeHostName, Address: "a1-yinvcez57-0-bvynoyawrhcg-kube-minion-fg5i4jwcc2yy.novalocal"},
	}

	if !reflect.DeepEqual(want, addrs) {
		t.Errorf("nodeAddresses returned incorrect value, want %v", want)
	}
}

func TestNodeAddressesPreferredOrderIPv6BeforeIPv4(t *testing.T) {
	srv, interfaces := buildNodeAddressesTestConfig()

	networkingOpts := NetworkingOpts{
		PublicNetworkName:     []string{"public"},
		PreferredAddressOrder: fmt.Sprintf("%s, %s, %s, %s", IPv6InternalOrderID, IPv6ExternalOrderID, IPv4InternalOrderID, IPv4ExternalOrderID),
	}

	addrs, err := nodeAddresses(srv, *interfaces, networkingOpts)
	if err != nil {
		t.Fatalf("nodeAddresses returned error: %v", err)
	}

	t.Logf("addresses are %v", addrs)

	want := []v1.NodeAddress{
		{Type: v1.NodeExternalIP, Address: "2001:4800:790e:510:be76:4eff:fe04:82a8"},
		{Type: v1.NodeExternalIP, Address: "2001:4800:780e:510:be76:4eff:fe04:84a8"},
		{Type: v1.NodeInternalIP, Address: "10.0.0.32"},
		{Type: v1.NodeInternalIP, Address: "10.0.0.31"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.36"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.35"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.99"},
		{Type: v1.NodeHostName, Address: "a1-yinvcez57-0-bvynoyawrhcg-kube-minion-fg5i4jwcc2yy.novalocal"},
	}

	if !reflect.DeepEqual(want, addrs) {
		t.Errorf("nodeAddresses returned incorrect value, want %v", want)
	}
}

func TestNodeAddressesPreferredOrderExternalBeforeInternalIPAddresses(t *testing.T) {
	srv, interfaces := buildNodeAddressesTestConfig()

	networkingOpts := NetworkingOpts{
		PublicNetworkName:     []string{"public"},
		PreferredAddressOrder: fmt.Sprintf("%s, %s, %s, %s", IPv4ExternalOrderID, IPv6ExternalOrderID, IPv4InternalOrderID, IPv6InternalOrderID),
	}

	addrs, err := nodeAddresses(srv, *interfaces, networkingOpts)
	if err != nil {
		t.Fatalf("nodeAddresses returned error: %v", err)
	}

	t.Logf("addresses are %v", addrs)

	want := []v1.NodeAddress{
		{Type: v1.NodeExternalIP, Address: "50.56.176.99"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.36"},
		{Type: v1.NodeExternalIP, Address: "50.56.176.35"},
		{Type: v1.NodeExternalIP, Address: "2001:4800:790e:510:be76:4eff:fe04:82a8"},
		{Type: v1.NodeExternalIP, Address: "2001:4800:780e:510:be76:4eff:fe04:84a8"},
		{Type: v1.NodeInternalIP, Address: "10.0.0.31"},
		{Type: v1.NodeInternalIP, Address: "10.0.0.32"},
		{Type: v1.NodeHostName, Address: "a1-yinvcez57-0-bvynoyawrhcg-kube-minion-fg5i4jwcc2yy.novalocal"},
	}

	if !reflect.DeepEqual(want, addrs) {
		t.Errorf("nodeAddresses returned incorrect value, want %v", want)
	}
}
