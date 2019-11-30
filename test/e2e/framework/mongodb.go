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
	"time"

	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

const (
	MongodbUser             = "root"
	MongodbPassword         = "root"
	MongodbCredentialSecret = "mongodb-credential-secret"
)

var (
	MongodbServiceName    = rand.WithUniqSuffix("test-svc-mongodb")
	MongodbDeploymentName = rand.WithUniqSuffix("test-mongodb-deploy")
)

// DeployMongodb will do:
//	- create service
//	- create deployment
//  - create credential secret
func (f *Framework) DeployMongodb() (*appcat.AppReference, error) {
	label := map[string]string{
		"app": rand.WithUniqSuffix("test-mongodb"),
	}

	srv := core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      MongodbServiceName,
		},
		Spec: core.ServiceSpec{
			Selector: label,
			Ports: []core.ServicePort{
				{
					Name:       "tcp",
					Protocol:   core.ProtocolTCP,
					Port:       27017,
					TargetPort: intstr.FromInt(27017),
				},
			},
		},
	}

	mongodbCont := core.Container{
		Name:            "mongo",
		Image:           "mongo",
		ImagePullPolicy: "IfNotPresent",
		Env: []core.EnvVar{
			{
				Name:  "MONGO_INITDB_ROOT_USERNAME",
				Value: MongodbUser,
			},
			{
				Name:  "MONGO_INITDB_ROOT_PASSWORD",
				Value: MongodbPassword,
			},
		},
		Ports: []core.ContainerPort{
			{
				Name:          "mongodb",
				Protocol:      core.ProtocolTCP,
				ContainerPort: 27017,
			},
		},
		VolumeMounts: []core.VolumeMount{
			{
				MountPath: "/data/db",
				Name:      "data",
			},
		},
	}

	mongodbDeploy := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.namespace,
			Name:      MongodbDeploymentName,
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
						mongodbCont,
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
		return nil, err
	}

	_, err = f.CreateDeployment(mongodbDeploy)
	if err != nil {
		return nil, err
	}

	Eventually(func() bool {
		if obj, err := f.KubeClient.AppsV1().Deployments(f.namespace).Get(mongodbDeploy.GetName(), metav1.GetOptions{}); err == nil {
			return *obj.Spec.Replicas == obj.Status.ReadyReplicas
		}
		return false
	}, timeOut, pollingInterval).Should(BeTrue())

	time.Sleep(10 * time.Second)

	sr := &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MongodbCredentialSecret,
			Namespace: f.namespace,
		},
		Data: map[string][]byte{
			"username": []byte(MongodbUser),
			"password": []byte(MongodbPassword),
		},
	}

	_, err = f.KubeClient.CoreV1().Secrets(f.namespace).Create(sr)
	if err != nil {
		return nil, err
	}

	appRef := &appcat.AppReference{
		Name:      rand.WithUniqSuffix("mongo-app"),
		Namespace: f.namespace,
	}

	err = f.CreateAppBinding(&appcat.AppBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRef.Name,
			Namespace: f.namespace,
		},
		Spec: appcat.AppBindingSpec{
			Secret: &core.LocalObjectReference{
				Name: MongodbCredentialSecret,
			},
			ClientConfig: appcat.ClientConfig{
				Service: &appcat.ServiceReference{
					Name:   MongodbServiceName,
					Scheme: "mongodb",
					Port:   27017,
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

func (f *Framework) DeleteMongodb() error {
	err := f.DeleteService(MongodbServiceName, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteSecret(MongodbCredentialSecret, f.namespace)
	if err != nil {
		return err
	}

	err = f.DeleteDeployment(MongodbDeploymentName, f.namespace)
	return err
}
