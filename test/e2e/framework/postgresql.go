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
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/appscode/go/crypto/rand"
	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"kmodules.xyz/client-go/tools/portforward"
)

var (
	postgresqlServiceName    = rand.WithUniqSuffix("test-svc-postgresql")
	postgresqlDeploymentName = rand.WithUniqSuffix("test-postgresql-deploy")
)

func (f *Framework) DeployPostgresSQL() (string, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-postgresql"),
	}

	srv := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      postgresqlServiceName,
		},
		Spec: core.ServiceSpec{
			Selector: label,
			Ports: []core.ServicePort{
				{
					Name:       "tcp",
					Protocol:   core.ProtocolTCP,
					Port:       5432,
					TargetPort: intstr.FromInt(5432),
				},
			},
		},
	}

	url := fmt.Sprintf("%s.%s.svc:5432", postgresqlServiceName, f.namespace)

	postgresqlCont := core.Container{
		Name:            "postgres",
		Image:           "postgres:9.6.2",
		ImagePullPolicy: "IfNotPresent",
		Env: []core.EnvVar{
			{
				Name:  "POSTGRES_USER",
				Value: "postgres",
			},
			{
				Name:  "POSTGRES_PASSWORD",
				Value: "root",
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
				ValueFrom: &core.EnvVarSource{
					FieldRef: &core.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Ports: []core.ContainerPort{
			{
				Name:          "postgresql",
				Protocol:      core.ProtocolTCP,
				ContainerPort: 5432,
			},
		},
		VolumeMounts: []core.VolumeMount{
			{
				MountPath: "/var/lib/postgresql/data/pgdata",
				Name:      "data",
				SubPath:   "postgresgl-db",
			},
		},
		ReadinessProbe: &core.Probe{
			Handler: core.Handler{
				Exec: &core.ExecAction{
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
			Name:      postgresqlDeploymentName,
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
						postgresqlCont,
					},
					Volumes: []core.Volume{
						{
							Name: "data",
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
		return "", err
	}

	_, err = f.CreateDeployment(postgresqlDeploy)
	if err != nil {
		return "", err
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1().Deployments(f.namespace).Get(postgresqlDeploy.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	time.Sleep(10 * time.Second)

	// create table
	pods, err := f.KubeClient.CoreV1().Pods(f.namespace).List(metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(label).String(),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to get postgresql pods")
	}
	if len(pods.Items) == 0 {
		return "", errors.New("no postgresql pod found")
	}

	err = f.setupPostgreSQL(pods.Items[0])
	if err != nil {
		return "", err
	}

	return url, nil
}

func (f *Framework) setupPostgreSQL(pod core.Pod) error {

	portFwd := portforward.NewTunnel(f.KubeClient.CoreV1().RESTClient(), f.ClientConfig, pod.GetNamespace(), pod.GetName(), 5432)
	defer portFwd.Close()

	err := portFwd.ForwardPort()
	if err != nil {
		return errors.Wrapf(err, "failed to port forward for pod(%s)", pod.GetName())
	}

	conn := fmt.Sprintf("postgres://postgres:root@localhost:%s/database?sslmode=disable", strconv.Itoa(portFwd.Local))

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return errors.Wrap(err, "failed to create postgres connection")
	}
	defer db.Close()

	stmt := `CREATE TABLE vault_kv_store (
		parent_path TEXT COLLATE "C" NOT NULL,
		path        TEXT COLLATE "C",
		key         TEXT COLLATE "C",
		value       BYTEA,
		CONSTRAINT pkey PRIMARY KEY (path, key)
	)`

	_, err = db.Exec(stmt)
	if err != nil {
		return err
	}

	stmt = `CREATE INDEX parent_path_idx ON vault_kv_store (parent_path)`
	_, err = db.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}

func (f *Framework) DeletePostgresSQL() error {
	err := f.DeleteService(postgresqlServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(postgresqlDeploymentName, f.namespace)
	return err
}
