package controller

import (
	"github.com/appscode/kubernetes-webhook-util/admission"
	hooks "github.com/appscode/kubernetes-webhook-util/admission/v1beta1"
	webhook "github.com/appscode/kubernetes-webhook-util/admission/v1beta1/workload"
	workload "github.com/appscode/kubernetes-webhook-util/workload/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c *VaultController) NewCronJobWebhook() hooks.AdmissionHook {
	return webhook.NewWorkloadWebhook(
		schema.GroupVersionResource{
			Group:    "admission.vault.soter.ac",
			Version:  "v1alpha1",
			Resource: "cronjobs",
		},
		"cronjob",
		"CronJob",
		nil,
		&admission.ResourceHandlerFuncs{
			CreateFunc: func(obj runtime.Object) (runtime.Object, error) {
				w := obj.(*workload.Workload)
				err := c.mutateWorkload(w)
				return w, err
			},
			UpdateFunc: func(oldObj, newObj runtime.Object) (runtime.Object, error) {
				w := newObj.(*workload.Workload)
				err := c.mutateWorkload(w)
				return w, err
			},
		},
	)
}
