// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"github.com/gardener/network-problem-detector/pkg/common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("utils.go validation functions", func() {
	Describe("validateNodeSlice", func() {
		It("should pass for valid IPv6 address", func() {
			nodes := []config.Node{{Hostname: "ip-10-180-20-164.eu-west-1.compute.internal", InternalIPsV6: []string{"2001:db8::1"}}}
			errList := config.ValidateNodeSlice(nodes, field.NewPath("nodes"))
			Expect(errList).To(BeEmpty())
		})
		It("should pass for valid node", func() {
			nodes := []config.Node{{Hostname: "shoot--garden--gcp-cilium-etcd-n2-2-8-z1-67db8-498zq", InternalIPs: []string{"10.0.0.1"}}}
			errList := config.ValidateNodeSlice(nodes, field.NewPath("nodes"))
			Expect(errList).To(BeEmpty())
		})
		It("should return error for invalid IPs", func() {
			nodes := []config.Node{{Hostname: "node1", InternalIPs: []string{"not-an-ip"}}}
			errList := config.ValidateNodeSlice(nodes, field.NewPath("nodes"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return error for an invalid hostname", func() {
			nodes := []config.Node{{Hostname: "invalid_host", InternalIPs: []string{"10.0.0.1"}}}
			errList := config.ValidateNodeSlice(nodes, field.NewPath("nodes"))
			Expect(errList).NotTo(BeEmpty())
		})
	})

	Describe("validatePodEndpointSlice", func() {
		It("should pass for valid pod endpoint", func() {
			endpoints := []config.PodEndpoint{{Nodename: "node", Podname: "valid-pod", PodIP: "10.128.0.1", Port: 1234}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).To(BeEmpty())
		})
		It("should pass for valid IPv6 address", func() {
			endpoints := []config.PodEndpoint{{Nodename: "node", Podname: "valid-pod", PodIP: "2001:db8::1", Port: 1234}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).To(BeEmpty())
		})
		It("should return error for invalid IPs", func() {
			endpoints := []config.PodEndpoint{{Nodename: "node", Podname: "invalid-pod", PodIP: "not-an-ip", Port: 1234}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return error for invalid podname", func() {
			endpoints := []config.PodEndpoint{{Nodename: "node", Podname: "invalid/pod", PodIP: "2001:db8::1", Port: 1234}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return error for invalid nodename", func() {
			endpoints := []config.PodEndpoint{{Nodename: "invalid+name", Podname: "valid-pod", PodIP: "2001:db8::1", Port: 1234}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return an error for invalid port number", func() {
			endpoints := []config.PodEndpoint{{Nodename: "node", Podname: "valid-pod", PodIP: "2001:db8::1", Port: -1}}
			errList := config.ValidatePodEndpointSlice(endpoints, field.NewPath("podEndpoints"))
			Expect(errList).NotTo(BeEmpty())
		})
	})
	Describe("validateEndpoint", func() {
		It("should pass for valid endpoint", func() {
			endpoint := &config.Endpoint{Hostname: "kubernetes.default.svc.cluster.local", IP: "1.2.3.4", Port: 443}
			errList := config.ValidateEndpoint(endpoint, field.NewPath("endpoint"))
			Expect(errList).To(BeEmpty())
		})
		It("should return error for invalid hostname", func() {
			endpoint := &config.Endpoint{Hostname: "invalid/name", IP: "2001:db8::1", Port: 1234}
			errList := config.ValidateEndpoint(endpoint, field.NewPath("endpoint"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return error for invalid IPs", func() {
			endpoint := &config.Endpoint{Hostname: "kubernetes.default.svc.cluster.local", IP: "not-an-ip", Port: 443}
			errList := config.ValidateEndpoint(endpoint, field.NewPath("endpoint"))
			Expect(errList).NotTo(BeEmpty())
		})
		It("should return an error for invalid port number", func() {
			endpoint := &config.Endpoint{Hostname: "kubernetes.default.svc.cluster.local", IP: "2001:db8::1", Port: -1}
			errList := config.ValidateEndpoint(endpoint, field.NewPath("endpoint"))
			Expect(errList).NotTo(BeEmpty())
		})
	})
})
