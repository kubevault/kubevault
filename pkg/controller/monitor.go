/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	api "kubevault.dev/apimachinery/apis/kubevault/v1alpha1"
	"kubevault.dev/operator/pkg/vault/exporter"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *VaultController) ensureStatsService(vs *api.VaultServer) (*core.Service, kutil.VerbType, error) {
	var statsSvc api.NamedServiceTemplateSpec

	for i := range vs.Spec.ServiceTemplates {
		temp := vs.Spec.ServiceTemplates[i]
		if temp.Alias == api.VaultServerServiceStats {
			statsSvc = temp
		}
	}

	meta := metav1.ObjectMeta{
		Name:      vs.StatsServiceName(),
		Namespace: vs.Namespace,
	}

	return core_util.CreateOrPatchService(context.TODO(), c.kubeClient, meta, func(in *core.Service) *core.Service {
		in.Labels = vs.StatsLabels()
		core_util.EnsureOwnerReference(in, metav1.NewControllerRef(vs, api.SchemeGroupVersion.WithKind(api.ResourceKindVaultServer)))

		in.Annotations = statsSvc.Annotations
		in.Spec.Selector = vs.OffshootSelectors()
		monSpec := vs.Spec.Monitor
		port := monSpec.Prometheus.Exporter.Port
		if port <= 0 {
			port = exporter.VaultExporterFetchMetricsPort
		}
		desired := []core.ServicePort{
			{
				Name:       exporter.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       port,
				TargetPort: intstr.FromInt(exporter.VaultExporterFetchMetricsPort),
			},
		}
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, desired)
		in.Spec.ClusterIP = statsSvc.Spec.ClusterIP
		in.Spec.Type = statsSvc.Spec.Type
		in.Spec.ExternalIPs = statsSvc.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = statsSvc.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = statsSvc.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = statsSvc.Spec.ExternalTrafficPolicy
		in.Spec.HealthCheckNodePort = statsSvc.Spec.HealthCheckNodePort
		in.Spec.SessionAffinityConfig = statsSvc.Spec.SessionAffinityConfig
		return in
	}, metav1.PatchOptions{})
}

func (c *VaultController) ensureStatsServiceDeleted(vs *api.VaultServer) error {
	klog.Infof("deleting stats service %s/%s", vs.Namespace, vs.StatsServiceName())
	err := c.kubeClient.CoreV1().Services(vs.Namespace).Delete(
		context.TODO(),
		vs.StatsServiceName(),
		metav1.DeleteOptions{})
	return errors.WithStack(err)
}

func (c *VaultController) newMonitorController(vs *api.VaultServer) (mona.Agent, error) {
	monitorSpec := vs.Spec.Monitor

	if monitorSpec == nil {
		return nil, errors.Errorf("MonitorSpec not found in %v", vs.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.kubeClient, c.promClient), nil
	}

	return nil, errors.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *VaultController) addOrUpdateMonitor(vs *api.VaultServer) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(vs)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	vs.Spec.Monitor.Prometheus.Exporter.Port = exporter.VaultExporterFetchMetricsPort

	return agent.CreateOrUpdate(vs.StatsService(), vs.Spec.Monitor)
}

func (c *VaultController) getOldAgent(vs *api.VaultServer) mona.Agent {
	service, err := c.kubeClient.CoreV1().Services(vs.Namespace).Get(context.TODO(), vs.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.kubeClient, c.promClient)
}

func (c *VaultController) setNewAgent(vs *api.VaultServer) error {
	service, err := c.kubeClient.CoreV1().Services(vs.Namespace).Get(context.TODO(), vs.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(context.TODO(), c.kubeClient, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(vs.Spec.Monitor.Agent),
		})
		return in
	}, metav1.PatchOptions{})
	return err
}

func (c *VaultController) manageMonitor(vs *api.VaultServer) error {
	oldAgent := c.getOldAgent(vs)
	if vs.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != vs.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(vs.StatsService()); err != nil {
				klog.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
			}
		}

		if _, err := c.addOrUpdateMonitor(vs); err != nil {
			return err
		}
		return c.setNewAgent(vs)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(vs.StatsService()); err != nil {
			klog.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
		}
	}
	return nil
}
