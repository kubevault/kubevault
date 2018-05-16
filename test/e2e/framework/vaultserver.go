package framework


import (
	"github.com/appscode/go/crypto/rand"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "github.com/soter/vault-operator/apis/vault/v1alpha1"
)

const (
	vaultImage        = "vault"
	vaultImageVersion = "0.10.0"
)

func (f *Invocation) VaultServer(node int32, bs api.BackendStorageSpec) *api.VaultServer {
	return &api.VaultServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("vault-test"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.VaultServerSpec{
			Nodes:     node,
			BaseImage: vaultImage,
			Version:   vaultImageVersion,
			BackendStorage: bs,
		},
	}
}

func (f *Framework) CreateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.VaultServerClient.VaultV1alpha1().VaultServers(obj.Namespace).Create(obj)
}

func (f *Framework) GetVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.VaultServerClient.VaultV1alpha1().VaultServers(obj.Namespace).Get(obj.Name, metav1.GetOptions{})
}

func (f *Framework) UpdateVaultServer(obj *api.VaultServer) (*api.VaultServer, error) {
	return f.VaultServerClient.VaultV1alpha1().VaultServers(obj.Namespace).Update(obj)
}

func (f *Framework) DeleteVaultServer(meta metav1.ObjectMeta) error {
	return f.VaultServerClient.VaultV1alpha1().VaultServers(meta.Namespace).Delete(meta.Name, deleteInBackground())
}

func (f *Framework) EventuallyVaultServer(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() *api.VaultServer {
		obj, err := f.VaultServerClient.VaultV1alpha1().VaultServers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return obj
	})
}
