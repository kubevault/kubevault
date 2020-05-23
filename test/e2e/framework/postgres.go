/*
Copyright The KubeVault Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	PostgresUser             = "postgres"
	PostgresPassword         = "root"
	PostgresCredentialSecret = "pg-cred-secret"
)

var (
	postgresServiceName    = rand.WithUniqSuffix("test-svc-postgresql")
	postgresDeploymentName = rand.WithUniqSuffix("test-postgresql-deploy")
)

// DeployPostgres will do:
//	- create service
//	- create deployment
//  - create credential secret
func (f *Framework) DeployPostgres() (*appcat.AppReference, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-postgresql"),
	}

	srv := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      postgresServiceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: label,
			Ports: []corev1.ServicePort{
				{
					Name:       "tcp",
					Protocol:   corev1.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
		},
	}

	postgresqlCont := corev1.Container{
		Name:            "postgres",
		Image:           "postgres:9.6.2",
		ImagePullPolicy: "IfNotPresent",
		Env: []corev1.EnvVar{
			{
				Name:  "POSTGRES_USER",
				Value: PostgresUser,
			},
			{
				Name:  "POSTGRES_PASSWORD",
				Value: PostgresPassword,
			},
			{
				Name:  "POSTGRES_DB",
				Value: "database",
			},
			{
				Name:  "PGDATA",
				Value: "/var/lib/postgresql/data/pgdata",
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
				Name:          "postgresql",
				Protocol:      corev1.ProtocolTCP,
				ContainerPort: 5432,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/var/lib/postgresql/data/pgdata",
				Name:      "data",
				SubPath:   "postgresgl-db",
			},
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						"exec pg_isready --host $POD_IP",
					},
				},
			},
			InitialDelaySeconds: 5,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
			FailureThreshold:    10,
		},
	}

	postgresqlDeploy := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      postgresDeploymentName,
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
						postgresqlCont,
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
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

	_, err = f.CreateDeployment(postgresqlDeploy)
	if err != nil {
		return nil, err
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1().Deployments(f.namespace).Get(context.TODO(), postgresqlDeploy.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		} else {
			fmt.Println(err)
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	time.Sleep(10 * time.Second)

	sr := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PostgresCredentialSecret,
			Namespace: f.namespace,
		},
		Data: map[string][]byte{
			"username": []byte(PostgresUser),
			"password": []byte(PostgresPassword),
		},
	}

	_, err = f.KubeClient.CoreV1().Secrets(f.namespace).Create(context.TODO(), sr, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	appRef := &appcat.AppReference{
		Name:      rand.WithUniqSuffix("postgres-app"),
		Namespace: f.namespace,
	}

	err = f.CreateAppBinding(&appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRef.Name,
			Namespace: f.namespace,
		},
		Spec: appcat.AppBindingSpec{
			Secret: &core.LocalObjectReference{
				Name: PostgresCredentialSecret,
			},
			ClientConfig: appcat.ClientConfig{
				Service: &appcat.ServiceReference{
					Name:   postgresServiceName,
					Scheme: "postgresql",
					Port:   5432,
					Path:   "/",
					Query:  "sslmode=disable",
				},
				InsecureSkipTLSVerify: true,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return appRef, nil
}

func (f *Framework) DeletePostgres() error {
	err := f.DeleteService(postgresServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteSecret(PostgresCredentialSecret, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(postgresDeploymentName, f.namespace)
	return err
}
