package v1beta1

import (
	"fmt"
	"strings"

	"github.com/appscode/kutil"
	"github.com/appscode/kutil/discovery"
	"github.com/evanphx/json-patch"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/api/admissionregistration/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ValidatingWebhookXray struct {
	config      *rest.Config
	webhookName string
	testObj     runtime.Object
	op          v1beta1.OperationType
	transform   func(_ runtime.Object)
}

func NewCreateValidatingWebhookXray(config *rest.Config, webhookName string, testObj runtime.Object) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:      config,
		webhookName: webhookName,
		testObj:     testObj,
		op:          v1beta1.Create,
		transform:   nil,
	}
}

func NewUpdateValidatingWebhookXray(config *rest.Config, webhookName string, testObj runtime.Object, transform func(_ runtime.Object)) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:      config,
		webhookName: webhookName,
		testObj:     testObj,
		op:          v1beta1.Update,
		transform:   transform,
	}
}

func NewDeleteValidatingWebhookXray(config *rest.Config, webhookName string, testObj runtime.Object, transform func(_ runtime.Object)) *ValidatingWebhookXray {
	return &ValidatingWebhookXray{
		config:      config,
		webhookName: webhookName,
		testObj:     testObj,
		op:          v1beta1.Delete,
		transform:   transform,
	}
}

var ErrMissingKind = errors.New("test object missing kind")
var ErrMissingVersion = errors.New("test object missing version")
var ErrInactiveWebhook = errors.New("webhook is inactive")

var bypassValidatingWebhookXray = false

func init() {
	pflag.BoolVar(&bypassValidatingWebhookXray, "bypass-validating-webhook-xray", bypassValidatingWebhookXray, "if true, bypasses validating webhook xray checks")
}

func (d ValidatingWebhookXray) IsActive() (bool, error) {
	kc, err := kubernetes.NewForConfig(d.config)
	if err != nil {
		return false, err
	}

	dc, err := dynamic.NewForConfig(d.config)
	if err != nil {
		return false, err
	}

	gvk := d.testObj.GetObjectKind().GroupVersionKind()
	if gvk.Version == "" {
		return false, ErrMissingVersion
	}
	if gvk.Kind == "" {
		return false, ErrMissingKind
	}
	glog.Infof("testing ValidatingWebhook %s using an object with GVK = %s", d.webhookName, gvk.String())

	gvr, err := discovery.ResourceForGVK(kc.Discovery(), gvk)
	if err != nil {
		return false, err
	}
	glog.Infof("testing ValidatingWebhook %s using an object with GVR = %s", d.webhookName, gvr.String())

	accessor, err := meta.Accessor(d.testObj)
	if err != nil {
		return false, err
	}

	var ri dynamic.ResourceInterface
	if accessor.GetNamespace() != "" {
		ri = dc.Resource(gvr).Namespace(accessor.GetNamespace())
	} else {
		ri = dc.Resource(gvr)
	}

	objJson, err := json.Marshal(d.testObj)
	if err != nil {
		return false, err
	}

	u := unstructured.Unstructured{}
	_, _, err = unstructured.UnstructuredJSONScheme.Decode(objJson, nil, &u)
	if err != nil {
		return false, err
	}

	if d.op == v1beta1.Create {
		_, err := ri.Create(&u)
		if kerr.IsForbidden(err) &&
			strings.HasPrefix(err.Error(), fmt.Sprintf(`admission webhook "%s" denied the request:`, d.webhookName)) {
			glog.Infof("failed to create invalid test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		err = ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return false, ErrInactiveWebhook
	} else if d.op == v1beta1.Update {
		_, err := ri.Create(&u)
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		mod := d.testObj.DeepCopyObject()
		d.transform(mod)
		modJson, err := json.Marshal(mod)
		if err != nil {
			return false, err
		}

		patch, err := jsonpatch.CreateMergePatch(objJson, modJson)
		if err != nil {
			return false, err
		}

		_, err = ri.Patch(accessor.GetName(), types.MergePatchType, patch)
		defer ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})

		if kerr.IsForbidden(err) &&
			strings.HasPrefix(err.Error(), fmt.Sprintf(`admission webhook "%s" denied the request:`, d.webhookName)) {
			glog.Infof("failed to update test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		return false, ErrInactiveWebhook
	} else if d.op == v1beta1.Delete {
		_, err := ri.Create(&u)
		if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		err = ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
		if kerr.IsForbidden(err) &&
			strings.HasPrefix(err.Error(), fmt.Sprintf(`admission webhook "%s" denied the request:`, d.webhookName)) {
			defer func() {
				// update to make it valid
				mod := d.testObj.DeepCopyObject()
				d.transform(mod)
				modJson, err := json.Marshal(mod)
				if err != nil {
					return
				}

				patch, err := jsonpatch.CreateMergePatch(objJson, modJson)
				if err != nil {
					return
				}

				ri.Patch(accessor.GetName(), types.MergePatchType, patch)

				// delete
				ri.Delete(accessor.GetName(), &metav1.DeleteOptions{})
			}()

			glog.Infof("failed to delete test object as expected with error: %s", err)
			return true, nil
		} else if kutil.IsRequestRetryable(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return false, ErrInactiveWebhook
	}

	return false, nil
}
