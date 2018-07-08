package framework

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	etcdServiceName    = rand.WithUniqSuffix("test-svc-etcd")
	etcdDeploymentName = rand.WithUniqSuffix("test-etcd-deploy")
)

func (f *Framework) DeployEtcd() (string, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-etcd"),
	}

	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      etcdServiceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: label,
			Ports: []corev1.ServicePort{
				{
					Name:       "client",
					Protocol:   corev1.ProtocolTCP,
					Port:       2379,
					TargetPort: intstr.FromInt(2379),
				},
				{
					Name:       "peer",
					Protocol:   corev1.ProtocolTCP,
					Port:       2380,
					TargetPort: intstr.FromInt(2380),
				},
			},
		},
	}

	clientUrl := fmt.Sprintf("http://%s.%s.svc:2379", etcdServiceName, f.namespace)
	peerUrl := fmt.Sprintf("http://%s.%s.svc:2380", etcdServiceName, f.namespace)

	etcdCont := corev1.Container{
		Name:  "etcd",
		Image: "quay.io/coreos/etcd:v3.2.13",
		Command: []string{
			"/usr/local/bin/etcd",
			"--data-dir=/var/etcd/data",
			"--name=$(MY_POD_NAME)",
			"--listen-peer-urls=http://0.0.0.0:2380",
			"--listen-client-urls=http://0.0.0.0:2379",
			"--initial-cluster-state=new",
			"--initial-cluster-token=12345",
			fmt.Sprintf("--initial-advertise-peer-urls=%s", peerUrl),
			fmt.Sprintf("--advertise-client-urls=%s", clientUrl),
			fmt.Sprintf("--initial-cluster=$(MY_POD_NAME)=%s", peerUrl),
		},
		Env: []corev1.EnvVar{
			{
				Name: "MY_POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name: "MY_POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
		},
	}

	etcdDeploy := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      etcdDeploymentName,
		},
		Spec: apps.DeploymentSpec{
			Replicas: func(i int32) *int32 { return &i }(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: label,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						etcdCont,
					},
				},
			},
		},
	}

	err := f.CreateService(srv)
	if err != nil {
		return "", err
	}

	_, err = f.CreateDeployment(etcdDeploy)
	if err != nil {
		return "", err
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1beta1().Deployments(f.namespace).Get(etcdDeploy.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	return fmt.Sprintf("http://%s.%s.svc:2379", srv.GetName(), srv.GetNamespace()), nil
}

func (f *Framework) DeleteEtcd() error {
	err := f.DeleteService(etcdServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(etcdDeploymentName, f.namespace)
	return err
}
