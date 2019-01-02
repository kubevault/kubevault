package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/exporter"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (c *VaultController) ensureStatsService(vs *api.VaultServer) (*core.Service, kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      vs.StatsServiceName(),
		Namespace: vs.Namespace,
	}

	return core_util.CreateOrPatchService(c.kubeClient, meta, func(in *core.Service) *core.Service {
		in.Labels = vs.StatsLabels()
		util.EnsureOwnerRefToObject(in, util.AsOwner(vs))

		in.Spec.Selector = vs.OffshootSelectors()
		monSpec := vs.Spec.Monitor
		port := monSpec.Prometheus.Port
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
		return in
	})
}

func (c *VaultController) ensureStatsServiceDeleted(vs *api.VaultServer) error {
	log.Infof("deleting stats service %s/%s", vs.Namespace, vs.StatsServiceName())
	err := c.kubeClient.CoreV1().Services(vs.Namespace).Delete(
		vs.StatsServiceName(),
		&metav1.DeleteOptions{},
	)
	return errors.WithStack(err)
}

func (c *VaultController) newMonitorController(vs *api.VaultServer) (mona.Agent, error) {
	monitorSpec := vs.Spec.Monitor

	fmt.Println(monitorSpec)
	if monitorSpec == nil {
		return nil, errors.Errorf("MonitorSpec not found in %v", vs.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.kubeClient, c.crdClient, c.promClient), nil
	}

	return nil, errors.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (c *VaultController) addOrUpdateMonitor(vs *api.VaultServer) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(vs)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	vs.Spec.Monitor.Prometheus.Port = exporter.VaultExporterFetchMetricsPort

	return agent.CreateOrUpdate(vs.StatsService(), vs.Spec.Monitor)
}

func (c *VaultController) getOldAgent(vs *api.VaultServer) mona.Agent {
	service, err := c.kubeClient.CoreV1().Services(vs.Namespace).Get(vs.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.kubeClient, c.crdClient, c.promClient)
}

func (c *VaultController) setNewAgent(vs *api.VaultServer) error {
	service, err := c.kubeClient.CoreV1().Services(vs.Namespace).Get(vs.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.kubeClient, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(vs.Spec.Monitor.Agent),
		},
		)

		return in
	})
	return err
}

func (c *VaultController) manageMonitor(vs *api.VaultServer) error {
	oldAgent := c.getOldAgent(vs)
	if vs.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != vs.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(vs.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
			}
		}

		if _, err := c.addOrUpdateMonitor(vs); err != nil {
			return err
		}
		return c.setNewAgent(vs)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(vs.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
		}
	}
	return nil
}
