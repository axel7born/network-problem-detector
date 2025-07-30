/*
 * SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"sigs.k8s.io/yaml"
)

var DisableShuffleForTesting = false

func LoadAgentConfig(configFile string) (*AgentConfig, error) {
	data, err := os.ReadFile(filepath.Clean(configFile))
	if err != nil {
		return nil, err
	}

	cfg := &AgentConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling %s failed: %w", configFile, err)
	}

	if errList := validateAgentConfig(cfg); len(errList) != 0 {
		return nil, fmt.Errorf("invalid network config: %w", errList.ToAggregate())
	}

	return cfg, nil
}

func validateAgentConfig(cfg *AgentConfig) field.ErrorList {

	var allErrs field.ErrorList

	if cfg.RetentionHours < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("retentionHours"), cfg.RetentionHours, "must be >= 0"))
	}
	if cfg.MaxPeerNodes < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("maxPeerNodes"), cfg.MaxPeerNodes, "must be >= 0"))
	}
	if cfg.K8sExporter != nil {
		if cfg.K8sExporter.MinFailingPeerNodeShare < 0.0 || cfg.K8sExporter.MinFailingPeerNodeShare > 1.0 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("k8sExporter.minFailingPeerNodeShare"), cfg.K8sExporter.MinFailingPeerNodeShare, "must be in [0.0, 1.0]"))
		}
	}

	return allErrs
}

// validateEndpoint validates a single Endpoint object.
func ValidateEndpoint(ep *Endpoint, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList
	if ep == nil {
		return errs
	}
	if ep.Hostname == "" {
		errs = append(errs, field.Required(basePath.Child("hostname"), "hostname is required"))
	} else {
		if dnsErrs := validation.IsDNS1123Subdomain(ep.Hostname); len(dnsErrs) > 0 {
			for _, err := range dnsErrs {
				errs = append(errs, field.Invalid(basePath.Child("hostname"), ep.Hostname, err))
			}
		}
	}
	if ep.IP == "" {
		errs = append(errs, field.Required(basePath.Child("ip"), "ip is required"))
	} else {
		if net.ParseIP(ep.IP) == nil {
			errs = append(errs, field.Invalid(basePath.Child("ip"), ep.IP, "must be a valid IP address"))
		}
	}
	if ep.Port < 0 || ep.Port > 65535 {
		errs = append(errs, field.Invalid(basePath.Child("port"), ep.Port, "must be in range 0-65535"))
	}
	return errs
}

// validateNodeSlice validates a slice of Node objects.
func ValidateNodeSlice(nodes []Node, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, n := range nodes {
		p := basePath.Index(i)
		if n.Hostname == "" {
			errs = append(errs, field.Required(p.Child("hostname"), "hostname is required"))
		} else {
			if dnsErrs := validation.IsDNS1123Subdomain(n.Hostname); len(dnsErrs) > 0 {
				for _, err := range dnsErrs {
					errs = append(errs, field.Invalid(p.Child("hostname"), n.Hostname, err))
				}
			}
		}
		if len(n.InternalIPs) == 0 && len(n.InternalIPsV6) == 0 {
			errs = append(errs, field.Required(p, "at least one internal IP (v4 or v6) is required"))
		}
		for j, ip := range n.InternalIPs {
			if net.ParseIP(ip) == nil {
				errs = append(errs, field.Invalid(p.Child("internalIPs").Index(j), ip, "must be a valid IP address"))
			}
		}
		for j, ip := range n.InternalIPsV6 {
			if net.ParseIP(ip) == nil {
				errs = append(errs, field.Invalid(p.Child("internalIPsV6").Index(j), ip, "must be a valid IP address"))
			}
		}
	}
	return errs
}

// validatePodEndpointSlice validates a slice of PodEndpoint objects.
func ValidatePodEndpointSlice(endpoints []PodEndpoint, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, ep := range endpoints {
		p := basePath.Index(i)
		if ep.Nodename == "" {
			errs = append(errs, field.Required(p.Child("nodename"), "nodename is required"))
		} else {
			if dnsErrs := validation.IsDNS1123Subdomain(ep.Nodename); len(dnsErrs) > 0 {
				for _, err := range dnsErrs {
					errs = append(errs, field.Invalid(p.Child("nodename"), ep.Nodename, err))
				}
			}
		}
		if ep.Podname == "" {
			errs = append(errs, field.Required(p.Child("podname"), "podname is required"))
		} else {
			if dnsErrs := validation.IsDNS1123Label(ep.Podname); len(dnsErrs) > 0 {
				for _, err := range dnsErrs {
					errs = append(errs, field.Invalid(p.Child("podname"), ep.Podname, err))
				}
			}
		}
		if ep.PodIP == "" {
			errs = append(errs, field.Required(p.Child("podIP"), "podIP is required"))
		} else {
			// Validate IP address format
			if net.ParseIP(ep.PodIP) == nil {
				errs = append(errs, field.Invalid(p.Child("podIP"), ep.PodIP, "must be a valid IP address"))
			}
		}
		if ep.Port < 0 || ep.Port > 65535 {
			errs = append(errs, field.Invalid(p.Child("port"), ep.Port, "must be in range 0-65535"))
		}
	}
	return errs
}

func validateClusterConfig(cfg *ClusterConfig) field.ErrorList {
	var allErrs field.ErrorList

	if cfg.NodeCount < 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("nodeCount"), cfg.NodeCount, "must be >= 0"))
	}
	if len(cfg.Nodes) > cfg.NodeCount {
		allErrs = append(allErrs, field.Invalid(field.NewPath("nodes"), len(cfg.Nodes), "cannot exceed nodeCount"))
	}

	allErrs = append(allErrs, ValidateNodeSlice(cfg.Nodes, field.NewPath("nodes"))...)

	allErrs = append(allErrs, ValidatePodEndpointSlice(cfg.PodEndpoints, field.NewPath("podEndpoints"))...)
	allErrs = append(allErrs, ValidatePodEndpointSlice(cfg.PodEndpointsV6, field.NewPath("podEndpointsV6"))...)

	allErrs = append(allErrs, ValidateEndpoint(cfg.InternalKubeAPIServer, field.NewPath("internalKubeAPIServer"))...)
	allErrs = append(allErrs, ValidateEndpoint(cfg.KubeAPIServer, field.NewPath("kubeAPIServer"))...)

	return allErrs
}

func LoadClusterConfig(configFile string) (*ClusterConfig, error) {
	data, err := os.ReadFile(filepath.Clean(configFile))
	if err != nil {
		return nil, err
	}

	cfg := &ClusterConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling %s failed: %w", configFile, err)
	}
	if errList := validateClusterConfig(cfg); len(errList) != 0 {
		return nil, fmt.Errorf("invalid cluster config: %w", errList.ToAggregate())
	}
	return cfg, nil
}

func CloneAndShuffle[T any](items []T) []T {
	if DisableShuffleForTesting {
		return items
	}
	if len(items) == 0 {
		return nil
	}
	clone := make([]T, len(items))
	copy(clone, items)
	rand.Shuffle(len(clone), func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
	return clone
}
