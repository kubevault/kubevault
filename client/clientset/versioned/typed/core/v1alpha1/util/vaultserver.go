package util

import (
	"encoding/json"

	"github.com/appscode/kutil"
	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	api "github.com/kube-vault/operator/apis/core/v1alpha1"
	cs "github.com/kube-vault/operator/client/clientset/versioned/typed/core/v1alpha1"
	"github.com/pkg/errors"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateOrPatchVaultServer(c cs.CoreV1alpha1Interface, meta metav1.ObjectMeta, transform func(alert *api.VaultServer) *api.VaultServer) (*api.VaultServer, kutil.VerbType, error) {
	cur, err := c.VaultServers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating VaultServer %s/%s.", meta.Namespace, meta.Name)
		out, err := c.VaultServers(meta.Namespace).Create(transform(&api.VaultServer{
			TypeMeta: metav1.TypeMeta{
				Kind:       api.ResourceKindVaultServer,
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchVaultServer(c, cur, transform)
}

func PatchVaultServer(c cs.CoreV1alpha1Interface, cur *api.VaultServer, transform func(*api.VaultServer) *api.VaultServer) (*api.VaultServer, kutil.VerbType, error) {
	return PatchVaultServerObject(c, cur, transform(cur.DeepCopy()))
}

func PatchVaultServerObject(c cs.CoreV1alpha1Interface, cur, mod *api.VaultServer) (*api.VaultServer, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching VaultServer %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.VaultServers(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateVaultServer(c cs.CoreV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.VaultServer) *api.VaultServer) (result *api.VaultServer, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.VaultServers(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.VaultServers(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update VaultServer %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = errors.Errorf("failed to update VaultServer %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func UpdateVaultServerStatus(c cs.CoreV1alpha1Interface, cur *api.VaultServer, transform func(*api.VaultServerStatus) *api.VaultServerStatus, useSubresource ...bool) (*api.VaultServer, error) {
	if len(useSubresource) > 1 {
		return nil, errors.Errorf("invalid value passed for useSubresource: %v", useSubresource)
	}

	mod := &api.VaultServer{
		TypeMeta:   cur.TypeMeta,
		ObjectMeta: cur.ObjectMeta,
		Spec:       cur.Spec,
		Status:     *transform(cur.Status.DeepCopy()),
	}

	if len(useSubresource) == 1 && useSubresource[0] {
		return c.VaultServers(cur.Namespace).UpdateStatus(mod)
	}

	out, _, err := PatchVaultServerObject(c, cur, mod)
	return out, err
}
