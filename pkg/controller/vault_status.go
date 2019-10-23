package controller

import (
	"context"
	"strconv"
	"time"

	api "kubevault.dev/operator/apis/kubevault/v1alpha1"
	cs_util "kubevault.dev/operator/client/clientset/versioned/typed/kubevault/v1alpha1/util"
	"kubevault.dev/operator/pkg/vault/util"

	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/client-go/tools/portforward"
)

// monitorAndUpdateStatus monitors the vault service and replicas statuses, and
// updates the status resource in the vault CR item.
func (c *VaultController) monitorAndUpdateStatus(ctx context.Context, vs *api.VaultServer) {
	tlsConfig := &vaultapi.TLSConfig{
		Insecure: true,
	}

	s := api.VaultServerStatus{
		Phase:       api.ClusterPhaseProcessing,
		ServiceName: vs.OffshootName(),
		ClientPort:  VaultClientPort,
		VaultStatus: api.VaultStatus{
			Standby: []string{},
			Sealed:  []string{},
		},
	}

	for {
		// Do not wait to update Phase ASAP.
		latest, err := c.updateVaultCRStatus(vs.Name, vs.Namespace, &s)
		if err != nil {
			glog.Errorf("vault status monitor: failed updating the status for the vault server %s: %v", vs.Name, err)
		}
		if latest != nil {
			vs = latest
			s = vs.Status
		}

		select {
		case err := <-ctx.Done():
			glog.Infof("vault status monitor: stop monitoring vault (%s/%s), reason: %v\n", vs.Namespace, vs.Name, err)
			return
		case <-time.After(5 * time.Second):
		}

		c.updateLocalVaultCRStatus(vs, &s, tlsConfig)
	}
}

// updateLocalVaultCRStatus updates local vault CR status by querying each vault pod's API.
func (c *VaultController) updateLocalVaultCRStatus(vs *api.VaultServer, s *api.VaultServerStatus, tlsConfig *vaultapi.TLSConfig) {
	name, namespace := vs.Name, vs.Namespace
	sel := vs.OffshootSelectors()

	// TODO : handle upgrades when pods from two replicaset can co-exist :(
	opt := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(sel).String()}

	version, err := c.extClient.CatalogV1alpha1().VaultServerVersions().Get(string(vs.Spec.Version), metav1.GetOptions{})
	if err != nil {
		glog.Errorf("vault status monitor: failed to get vault server version(%s): %v", vs.Spec.Version, err)
		return
	}

	pods, err := c.kubeClient.CoreV1().Pods(namespace).List(opt)
	if err != nil {
		glog.Errorf("vault status monitor: failed to update vault replica status: failed listing pods for the vault server (%s.%s): %v", namespace, name, err)
		return
	}

	if len(pods.Items) == 0 {
		glog.Errorf("vault status monitor: for the vault server (%s.%s): no pods found", namespace, name)
		return
	}

	activeNode := ""
	sealNodes := []string{}
	unsealNodes := []string{}
	standByNodes := []string{}
	updated := []string{}
	initiated := false
	// If it can't talk to any vault pod, we are not going to change the status.
	changed := false

	for _, p := range pods.Items {
		// If a pod is Terminating, it is still Running but has no IP.
		if p.Status.Phase != corev1.PodRunning || p.DeletionTimestamp != nil {
			continue
		}

		hr, err := c.getVaultStatus(&p, tlsConfig)
		if err != nil {
			glog.Error("vault status monitor:", err)
			continue
		}

		changed = true

		if p.Spec.Containers[0].Image == version.Spec.Vault.Image {
			updated = append(updated, p.Name)
		}

		if hr.Initialized && !hr.Sealed && !hr.Standby {
			activeNode = p.Name
		}
		if hr.Initialized && !hr.Sealed && hr.Standby {
			standByNodes = append(standByNodes, p.Name)
		}
		if hr.Sealed {
			sealNodes = append(sealNodes, p.Name)
		} else {
			unsealNodes = append(unsealNodes, p.Name)
		}
		if hr.Initialized {
			initiated = true
		}
	}

	if !changed {
		return
	}

	s.VaultStatus.Active = activeNode
	s.VaultStatus.Standby = standByNodes
	s.VaultStatus.Sealed = sealNodes
	s.VaultStatus.Unsealed = unsealNodes
	s.Initialized = initiated
	s.UpdatedNodes = updated
	if !s.Initialized {
		s.Phase = api.ClusterPhaseUnInitialized
	} else if activeNode != "" {
		// if there is an active vault node, then vault is ready to receive request
		s.Phase = api.ClusterPhaseRunning
	} else if len(sealNodes) > 0 {
		s.Phase = api.ClusterPhaseSealed
	}
}

// updateVaultCRStatus updates the status field of the Vault CR.
func (c *VaultController) updateVaultCRStatus(name, namespace string, status *api.VaultServerStatus) (*api.VaultServer, error) {
	vault, err := c.extClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		key := namespace + "/" + name
		if ctxWithcancel, ok := c.ctxCancels[key]; ok {
			ctxWithcancel.Cancel()
			delete(c.ctxCancels, key)
		}
		return nil, err
	} else if err != nil {
		return nil, err
	}

	// TODO : flag for useSubresource?
	vault, err = cs_util.UpdateVaultServerStatus(c.extClient.KubevaultV1alpha1(), vault, func(s *api.VaultServerStatus) *api.VaultServerStatus {
		s.VaultStatus.Active = status.VaultStatus.Active
		s.VaultStatus.Standby = status.VaultStatus.Standby
		s.VaultStatus.Sealed = status.VaultStatus.Sealed
		s.VaultStatus.Unsealed = status.VaultStatus.Unsealed
		s.Initialized = status.Initialized
		s.UpdatedNodes = status.UpdatedNodes
		s.Phase = status.Phase
		s.ServiceName = status.ServiceName
		s.ClientPort = status.ClientPort
		return s
	})
	return vault, err
}

func (c *VaultController) getVaultStatus(p *corev1.Pod, tlsConfig *vaultapi.TLSConfig) (*vaultapi.HealthResponse, error) {
	// podAddr contains pod access url
	// PodDNSName is reachable if operator running in cluster mode
	podAddr := util.PodDNSName(*p)
	// vault server pod use port 8200
	podPort := "8200"

	if !meta_util.PossiblyInCluster() {
		// if not incluster mode, use port forwarding to access pod

		portFwd := portforward.NewTunnel(c.kubeClient.CoreV1().RESTClient(), c.clientConfig, p.Namespace, p.Name, 8200)
		defer portFwd.Close()

		err := portFwd.ForwardPort()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get vault pod status: port forward failed for pod (%s/%s).", p.Namespace, p.Name)
		}

		podAddr = "localhost"
		podPort = strconv.Itoa(portFwd.Local)
	}

	vaultClient, err := util.NewVaultClient(podAddr, podPort, tlsConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get vault pod status: failed creating client for the vault pod (%s/%s).", p.Namespace, p.Name)
	}

	hr, err := vaultClient.Sys().Health()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get vault pod status: failed requesting health info for the vault pod (%s/%s).", p.Namespace, p.Name)
	}
	return hr, nil
}
