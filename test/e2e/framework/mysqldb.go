package framework

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	MysqlUser             = "root"
	MysqlPassword         = "root"
	MysqlCredentialSecret = "mysql-db-credential-secret"
)

var (
	mysqlDBServiceName    = rand.WithUniqSuffix("test-svc-mysql")
	mysqlDBDeploymentName = rand.WithUniqSuffix("test-mysql-deploy")
)

// DeployMysql will do:
//	- create service
//	- create deployment
//  - create credential secret
func (f *Framework) DeployMysql() (*appcat.AppReference, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-mysql"),
	}

	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      mysqlDBServiceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: label,
			Ports: []corev1.ServicePort{
				{
					Name:       "tcp",
					Protocol:   corev1.ProtocolTCP,
					Port:       3306,
					TargetPort: intstr.FromInt(3306),
				},
			},
		},
	}

	mysqlCont := corev1.Container{
		Name:            "mysql",
		Image:           "mysql:5.7",
		ImagePullPolicy: "IfNotPresent",
		Env: []corev1.EnvVar{
			{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: MysqlPassword,
			},
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "mysql",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: 3306,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/var/lib/mysql",
				Name:      "data",
			},
			{
				MountPath: "/etc/mysql",
				Name:      "config",
			},
		},
	}

	mysqlDeploy := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      mysqlDBDeploymentName,
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
						mysqlCont,
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	err := f.CreateService(srv)
	if err != nil {
		return nil, err
	}

	_, err = f.CreateDeployment(mysqlDeploy)
	if err != nil {
		return nil, err
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1beta1().Deployments(f.namespace).Get(mysqlDeploy.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	time.Sleep(10 * time.Second)

	sr := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MysqlCredentialSecret,
			Namespace: f.namespace,
		},
		Data: map[string][]byte{
			"username": []byte(MysqlUser),
			"password": []byte(MysqlPassword),
		},
	}

	_, err = f.KubeClient.CoreV1().Secrets(f.namespace).Create(sr)
	if err != nil {
		return nil, err
	}

	appRef := &appcat.AppReference{
		Name:      rand.WithUniqSuffix("mysql-app"),
		Namespace: f.namespace,
	}

	err = f.CreateAppBinding(&appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRef.Name,
			Namespace: f.namespace,
		},
		Spec: appcat.AppBindingSpec{
			Secret: &core.LocalObjectReference{
				Name: MysqlCredentialSecret,
			},
			ClientConfig: appcat.ClientConfig{
				URL:                   types.StringP(fmt.Sprintf("tcp(%s.%s.svc:3306)/", mysqlDBServiceName, f.namespace)),
				InsecureSkipTLSVerify: true,
			},
			Parameters: &runtime.RawExtension{
				Raw: []byte(`{
   "allowedRoles":"*",
   "pluginName":"mysql-rds-database-plugin"
}`),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return appRef, nil
}

func (f *Framework) DeleteMysql() error {
	err := f.DeleteService(mysqlDBServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteSecret(MysqlCredentialSecret, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(mysqlDBDeploymentName, f.namespace)
	return err
}
