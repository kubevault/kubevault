package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/kutil"
	core_util "github.com/appscode/kutil/core/v1"
	meta_util "github.com/appscode/kutil/meta"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	"github.com/kubevault/operator/pkg/vault/exporter"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
)

func (v *VaultController) newMonitorController(vaultserver *api.VaultServer) (mona.Agent, error) {
	monitorSpec := vaultserver.Spec.Monitor

	fmt.Println(monitorSpec)
	if monitorSpec == nil {
		return nil, errors.Errorf("MonitorSpec not found in %v", vaultserver.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, v.kubeClient, v.crdClient, v.promClient), nil
	}

	return nil, errors.Errorf("monitoring controller not found for %v", monitorSpec)
}

func (v *VaultController) addOrUpdateMonitor(vaultServer *api.VaultServer) (kutil.VerbType, error) {
	agent, err := v.newMonitorController(vaultServer)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	vaultServer.Spec.Monitor.Prometheus.Port = exporter.VaultExporterFetchMetricsPort

	return agent.CreateOrUpdate(vaultServer.StatsService(), vaultServer.Spec.Monitor)
}

func (v *VaultController) getOldAgent(vaultserver *api.VaultServer) mona.Agent {
	service, err := v.kubeClient.CoreV1().Services(vaultserver.Namespace).Get(vaultserver.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), v.kubeClient, v.crdClient, v.promClient)
}

func (v *VaultController) setNewAgent(vaultserver *api.VaultServer) error {
	service, err := v.kubeClient.CoreV1().Services(vaultserver.Namespace).Get(vaultserver.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(v.kubeClient, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(vaultserver.Spec.Monitor.Agent),
		},
		)

		return in
	})
	return err
}

func (c *VaultController) manageMonitor(vaultserver *api.VaultServer) error {
	oldAgent := c.getOldAgent(vaultserver)
	if vaultserver.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != vaultserver.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(vaultserver.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
			}
		}

		if _, err := c.addOrUpdateMonitor(vaultserver); err != nil {
			return err
		}
		return c.setNewAgent(vaultserver)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(vaultserver.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %v", err.Error())
		}
	}
	return nil
}
