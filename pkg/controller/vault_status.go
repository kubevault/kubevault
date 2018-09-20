package controller

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	meta_util "github.com/appscode/kutil/meta"
	"github.com/appscode/kutil/tools/portforward"
	"github.com/golang/glog"
	vaultapi "github.com/hashicorp/vault/api"
	api "github.com/kubevault/operator/apis/kubevault/v1alpha1"
	patchutil "github.com/kubevault/operator/client/clientset/versioned/typed/kubevault/v1alpha1/util"
	"github.com/kubevault/operator/pkg/vault/util"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	caFileDir = ".pki"
)

// monitorAndUpdateStatus monitors the vault service and replicas statuses, and
// updates the status resource in the vault CR item.
func (c *VaultController) monitorAndUpdateStatus(ctx context.Context, v *api.VaultServer) {
	var tlsConfig *vaultapi.TLSConfig
	tlsSecretName := v.TLSSecretName()

	appFs := afero.NewOsFs()
	appFs.Mkdir(caFileDir, 0777)
	defer appFs.RemoveAll(caFileDir)

	s := api.VaultServerStatus{
		Phase:       api.ClusterPhaseRunning,
		ServiceName: v.OffshootName(),
		ClientPort:  VaultPort,
		VaultStatus: api.VaultStatus{
			Standby: []string{},
			Sealed:  []string{},
		},
	}

	for {
		// Do not wait to update Phase ASAP.
		latest, err := c.updateVaultCRStatus(ctx, v.Name, v.Namespace, &s)
		if err != nil {
			glog.Errorf("vault status monitor: failed updating the status for the vault server %s: %v", v.Name, err)
		}
		if latest != nil {
			v = latest
		}

		select {
		case err := <-ctx.Done():
			glog.Infof("vault status monitor: stop monitoring vault (%s/%s), reason: %v\n", v.Namespace, v.Name, err)
			return
		case <-time.After(10 * time.Second):
		}

		if tlsConfig == nil {
			if v.Spec.TLS != nil {
				if v.Spec.TLS.TLSSecret != "" {
					tlsSecretName = v.Spec.TLS.TLSSecret
				}
			}
			se, err := c.kubeClient.CoreV1().Secrets(v.Namespace).Get(tlsSecretName, metav1.GetOptions{})
			if err != nil {
				glog.Errorf("vault status monitor: failed get secret `%v`", tlsSecretName)
			}

			caFile := filepath.Join(caFileDir, CaCertName)

			afero.WriteFile(appFs, caFile, se.Data[CaCertName], 0777)

			tlsConfig = &vaultapi.TLSConfig{
				CACert: caFile,
			}

			if err != nil {
				glog.Errorf("vault status monitor: failed to read TLS config for vault client: %v", err)
				continue
			}
		}
		c.updateLocalVaultCRStatus(ctx, v, &s, tlsConfig)
	}
}

// updateLocalVaultCRStatus updates local vault CR status by querying each vault pod's API.
func (c *VaultController) updateLocalVaultCRStatus(ctx context.Context, v *api.VaultServer, s *api.VaultServerStatus, tlsConfig *vaultapi.TLSConfig) {
	name, namespace := v.Name, v.Namespace
	sel := v.OffshootSelectors()

	// TODO : handle upgrades when pods from two replicaset can co-exist :(
	opt := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(sel).String()}

	version, err := c.extClient.KubevaultV1alpha1().VaultServerVersions().Get(string(v.Spec.Version), metav1.GetOptions{})
	if err != nil {
		glog.Errorf("vault status monitor: failed to get vault server version(%s): %v", v.Spec.Version, err)
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
}

// updateVaultCRStatus updates the status field of the Vault CR.
func (c *VaultController) updateVaultCRStatus(ctx context.Context, name, namespace string, status *api.VaultServerStatus) (*api.VaultServer, error) {
	vault, err := c.extClient.KubevaultV1alpha1().VaultServers(namespace).Get(name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		key := namespace + "/" + name
		if cancel, ok := c.ctxCancels[key]; ok {
			cancel()
			delete(c.ctxCancels, key)
		}
		return nil, err
	} else if err != nil {
		return nil, err
	}

	// TODO : flag for useSubresource?
	vault, err = patchutil.UpdateVaultServerStatus(c.extClient.KubevaultV1alpha1(), vault, func(s *api.VaultServerStatus) *api.VaultServerStatus {
		*s = *status
		return s
	}, api.EnableStatusSubresource)
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

		portFwd := portforward.NewTunnel(c.kubeClient.CoreV1().RESTClient(), c.restConfig, p.Namespace, p.Name, 8200)
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
