// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package digitalocean

import (
	"fmt"
	"os"

	"github.com/kubicorn/kubicorn/apis/cluster"
	"github.com/kubicorn/kubicorn/pkg/kubeadm"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewUbuntuCluster creates a basic Digitalocean cluster profile, to bootstrap Kubernetes.
func NewUbuntuCluster(name string) *cluster.Cluster {

	controlPlaneProviderConfig := &cluster.ControlPlaneProviderConfig{
		Cloud:    cluster.CloudDigitalOcean,
		Location: "sfo2",
		SSH: &cluster.SSH{
			PublicKeyPath: "~/.ssh/id_rsa.pub",
			User:          "root",
		},
		KubernetesAPI: &cluster.KubernetesAPI{
			Port: "443",
		},
		Values: &cluster.Values{
			ItemMap: map[string]string{
				"INJECTEDTOKEN": kubeadm.GetRandomToken(),
			},
		},
	}
	machineSetsProviderConfigs := []*cluster.MachineProviderConfig{
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeMaster,
				Name:     fmt.Sprintf("%s-master", name),
				MaxCount: 1,
				Image:    "ubuntu-16-04-x64",
				Size:     "s-2vcpu-2gb",
				BootstrapScripts: []string{
					"bootstrap/digitalocean_k8s_ubuntu_16.04_master.sh",
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s-master", name),
						IngressRules: []*cluster.IngressRule{
							{
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "443",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-node", name),
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-node", name),
								IngressProtocol: "udp",
							},
						},
						EgressRules: []*cluster.EgressRule{
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "tcp",
							},
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "udp",
							},
						},
					},
				},
			},
		},
		{
			ServerPool: &cluster.ServerPool{
				Type:     cluster.ServerPoolTypeNode,
				Name:     fmt.Sprintf("%s-node", name),
				MaxCount: 2,
				Image:    "ubuntu-16-04-x64",
				Size:     "s-1vcpu-2gb",
				BootstrapScripts: []string{
					"bootstrap/digitalocean_k8s_ubuntu_16.04_node.sh",
				},
				Firewalls: []*cluster.Firewall{
					{
						Name: fmt.Sprintf("%s-node", name),
						IngressRules: []*cluster.IngressRule{
							{
								IngressToPort:   "22",
								IngressSource:   "0.0.0.0/0",
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-master", name),
								IngressProtocol: "tcp",
							},
							{
								IngressToPort:   "all",
								IngressSource:   fmt.Sprintf("%s-master", name),
								IngressProtocol: "udp",
							},
						},
						EgressRules: []*cluster.EgressRule{
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "tcp",
							},
							{
								EgressToPort:      "all", // By default all egress from VM
								EgressDestination: "0.0.0.0/0",
								EgressProtocol:    "udp",
							},
						},
					},
				},
			},
		},
	}
	c := cluster.NewCluster(name)
	c.SetProviderConfig(controlPlaneProviderConfig)
	c.NewMachineSetsFromProviderConfigs(machineSetsProviderConfigs)

	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "digitalocean",
			Namespace: "kube-system",
		},
		StringData: map[string]string{"access-token": string(os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))},
	}
	c.APITokenSecret = secret

	return c
}
