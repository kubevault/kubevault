package controller

import (
	"context"
	"fmt"
	"testing"

	api "github.com/kubevault/operator/apis/core/v1alpha1"
	cfake "github.com/kubevault/operator/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

type vaultFake struct {
	sr                *corev1.Secret
	cm                *corev1.ConfigMap
	dp                *appsv1.Deployment
	sa                *corev1.ServiceAccount
	svc               *corev1.Service
	roles             []rbac.Role
	pt                *corev1.PodTemplateSpec
	cnt               corev1.Container
	ErrInGetServerTLS bool
	ErrInGetConfig    bool
	ErrInApply        bool
}

func (v *vaultFake) GetServerTLS() (*corev1.Secret, error) {
	if v.ErrInGetServerTLS {
		return nil, fmt.Errorf("error")
	}
	return v.sr, nil
}
func (v *vaultFake) GetConfig() (*corev1.ConfigMap, error) {
	if v.ErrInGetConfig {
		return nil, fmt.Errorf("error")
	}
	return v.cm, nil
}
func (v *vaultFake) Apply(pt *corev1.PodTemplateSpec) error {
	if v.ErrInApply {
		return fmt.Errorf("error")
	}
	return nil
}
func (v *vaultFake) GetService() *corev1.Service {
	return v.svc
}
func (v *vaultFake) GetDeployment(pt *corev1.PodTemplateSpec) *appsv1.Deployment {
	return v.dp
}
func (v *vaultFake) GetServiceAccount() *corev1.ServiceAccount {
	return v.sa
}
func (v *vaultFake) GetRBACRoles() []rbac.Role {
	return v.roles
}
func (v *vaultFake) GetPodTemplate(c corev1.Container, saName string) *corev1.PodTemplateSpec {
	return v.pt
}
func (v *vaultFake) GetContainer() corev1.Container {
	return v.cnt
}

func TestReconcileVault(t *testing.T) {
	vfk := vaultFake{
		sr: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sr-test",
				Namespace: "test",
			},
		},
		cm: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-test",
				Namespace: "test",
			},
		},
		sa: &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-test",
				Namespace: "test",
			},
		},
		cnt: corev1.Container{},
		pt:  &corev1.PodTemplateSpec{},
		dp: &appsv1.Deployment{
			ObjectMeta: getVaultObjectMeta(1),
		},
		svc: &corev1.Service{
			ObjectMeta: getVaultObjectMeta(1),
		},
		roles: []rbac.Role{},
	}

	testData := []struct {
		name      string
		vs        *api.VaultServer
		vfake     *vaultFake
		expectErr bool
	}{
		{
			name: "no error",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},

			vfake:     &vfk,
			expectErr: false,
		},
		{
			name: "failed to create vault tls",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},

			vfake:     func(v vaultFake) *vaultFake { v.ErrInGetServerTLS = true; return &v }(vfk),
			expectErr: true,
		},
		{
			name: "failed to create vault config",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},

			vfake:     func(v vaultFake) *vaultFake { v.ErrInGetConfig = true; return &v }(vfk),
			expectErr: true,
		},
		{
			name: "failed to deploy vault",
			vs: &api.VaultServer{
				ObjectMeta: getVaultObjectMeta(1),
			},

			vfake:     func(v vaultFake) *vaultFake { v.ErrInApply = true; return &v }(vfk),
			expectErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			vaultCtrl := VaultController{
				kubeClient: kfake.NewSimpleClientset(),
				recorder:   record.NewFakeRecorder(0),
				ctxCancels: map[string]context.CancelFunc{},
				extClient:  cfake.NewSimpleClientset(),
			}

			// to ignore monitorAndUpdateStatus
			vaultCtrl.ctxCancels[test.vs.Name] = func() {}

			err := vaultCtrl.reconcileVault(test.vs, test.vfake)
			if test.expectErr {
				assert.NotNil(t, err, "error must be non-empty")
			} else {
				assert.Nil(t, err, "error must be nil")

				_, err := vaultCtrl.kubeClient.AppsV1().Deployments(test.vs.Namespace).Get(test.vs.Name, metav1.GetOptions{})
				assert.Nil(t, err, "deployment for vaultserver should exist")
			}
		})
	}
}

func TestDeployVault(t *testing.T) {
	vfk := vaultFake{
		sa: &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sa-test",
				Namespace: "test",
			},
		},
		cnt: corev1.Container{},
		pt:  &corev1.PodTemplateSpec{},
		dp: &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dp-test",
				Namespace: "test",
			},
		},
		svc: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-test",
				Namespace: "test",
			},
		},
		roles:      []rbac.Role{},
		ErrInApply: false,
	}
	testData := []struct {
		name      string
		vfake     *vaultFake
		vs        *api.VaultServer
		expectErr bool
	}{
		{
			name:  "no error",
			vfake: &vfk,
			vs: &api.VaultServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-test",
					Namespace: "test",
				},
			},
			expectErr: false,
		},
		{
			name:  "expected error",
			vfake: func(v vaultFake) *vaultFake { v.ErrInApply = true; return &v }(vfk),
			vs: &api.VaultServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "svc-test",
					Namespace: "test",
				},
			},
			expectErr: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			vaultCtrl := VaultController{
				kubeClient: kfake.NewSimpleClientset(),
			}

			err := vaultCtrl.DeployVault(test.vs, test.vfake)
			if test.expectErr {
				assert.NotNil(t, err, "error must be non-empty")
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestCreateRoleAndRoleBinding(t *testing.T) {
	vaultCtrl := VaultController{
		kubeClient: kfake.NewSimpleClientset(),
	}

	demoRole := rbac.Role{
		Rules: []rbac.PolicyRule{
			{
				APIGroups: []string{corev1.GroupName},
				Resources: []string{"secret"},
				Verbs:     []string{"*"},
			},
		},
	}

	vs := &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "try",
			Namespace: "default",
		},
	}

	testData := []struct {
		testName           string
		preCreatedRole     []rbac.Role
		roles              []rbac.Role
		expectErr          bool
		expectRoles        []string
		expectRoleBindings []string
	}{
		{
			"create 2 rbac role and rolebinding",
			nil,
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test1"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
				func(r *rbac.Role) rbac.Role { r.SetName("test2"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			false,
			[]string{"test1", "test2"},
			[]string{"test1", "test2"},
		},
		{
			"create 1 rbac role and rolebinding, but role already exists",
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			[]rbac.Role{
				func(r *rbac.Role) rbac.Role { r.SetName("test3"); r.SetNamespace(vs.Namespace); return *r }(&demoRole),
			},
			false,
			[]string{"test3"},
			[]string{"test3"},
		},
	}

	for _, test := range testData {
		t.Run(test.testName, func(t *testing.T) {
			for _, r := range test.preCreatedRole {
				_, err := vaultCtrl.kubeClient.RbacV1().Roles(vs.Namespace).Create(&r)
				assert.Nil(t, err)
			}

			err := ensureRoleAndRoleBinding(vaultCtrl.kubeClient, vs, test.roles, "try")
			if test.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			for _, r := range test.expectRoles {
				_, err := vaultCtrl.kubeClient.RbacV1().Roles(vs.Namespace).Get(r, metav1.GetOptions{})
				assert.Nil(t, err, fmt.Sprintf("role(%s) should exists", r))
			}

			for _, rb := range test.expectRoleBindings {
				_, err := vaultCtrl.kubeClient.RbacV1().RoleBindings(vs.Namespace).Get(rb, metav1.GetOptions{})
				assert.Nil(t, err, fmt.Sprintf("rolebinding (%s) should exists", rb))
			}
		})
	}
}
