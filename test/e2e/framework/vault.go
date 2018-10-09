package framework

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

var (
	vaultServiceName    = rand.WithUniqSuffix("test-svc-vault")
	vaultDeploymentName = rand.WithUniqSuffix("test-vault-deploy")
)

const (
	nodePort         = 30088
	VaultTokenSecret = "vault-token"
)

// DeployVault will do
//	- create service
//	- create deployment
//	- create vault token secret
func (f *Framework) DeployVault() (*appcat.AppReference, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-vault"),
	}

	srv := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      vaultServiceName,
		},
		Spec: core.ServiceSpec{
			Selector: label,
			Ports: []core.ServicePort{
				{
					Name:     "http",
					Protocol: core.ProtocolTCP,
					Port:     8200,
					NodePort: nodePort,
				},
			},
			Type: core.ServiceTypeNodePort,
		},
	}

	err := f.CreateService(srv)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create service(%s/%s)", srv.Namespace, srv.Name)
	}

	d := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      vaultDeploymentName,
		},
		Spec: apps.DeploymentSpec{
			Replicas: func(i int32) *int32 { return &i }(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "vault",
							Image: "vault:0.10.3",
							Args: []string{
								"server",
								"-dev",
								"-dev-root-token-id=root",
							},
							Ports: []core.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8200,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = f.CreateDeployment(d)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create deployment(%s/%s)", d.Namespace, d.Name)
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1beta1().Deployments(f.namespace).Get(d.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	time.Sleep(10 * time.Second)

	sr := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VaultTokenSecret,
			Namespace: f.namespace,
		},
		Data: map[string][]byte{
			"token": []byte("root"),
		},
	}

	_, err = f.KubeClient.CoreV1().Secrets(f.namespace).Create(sr)
	if err != nil {
		return nil, err
	}

	nodePortIP, err := f.getNodePortIP(label)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d", nodePortIP, nodePort)

	appRef := &appcat.AppReference{
		Name:      rand.WithUniqSuffix("vault-ref"),
		Namespace: f.namespace,
	}

	err = f.CreateAppBinding(&appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRef.Name,
			Namespace: f.namespace,
		},
		Spec: appcat.AppBindingSpec{
			Secret: &core.LocalObjectReference{
				VaultTokenSecret,
			},
			ClientConfig: appcat.ClientConfig{
				URL: &url,
				InsecureSkipTLSVerify: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return appRef, nil
}

func (f *Framework) DeleteVault() error {
	err := f.DeleteService(vaultServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteSecret(VaultTokenSecret, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteAppBinding(f.VaultAppRef.Name, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(vaultDeploymentName, f.namespace)
	return err
}

func (f *Framework) getNodePortIP(label map[string]string) (string, error) {
	pods, err := f.KubeClient.CoreV1().Pods(f.namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(label).String(),
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) != 1 {
		return "", errors.New("number of vault pods is not 1")
	}

	for _, p := range pods.Items {
		node, err := f.KubeClient.CoreV1().Nodes().Get(p.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		for _, addr := range node.Status.Addresses {
			if addr.Type == core.NodeExternalIP {
				return addr.Address, nil
			}
		}

		if node.Name == "minikube" {
			return getMinikubeIP()
		}
	}

	return "", errors.New("no ip found")
}

func getMinikubeIP() (string, error) {
	ip, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get minikube ip")
	}

	return strings.TrimSpace(string(ip)), err
}
