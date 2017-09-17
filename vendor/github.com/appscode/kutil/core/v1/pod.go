package v1

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/appscode/kutil"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func EnsurePod(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.Pod) *apiv1.Pod) (*apiv1.Pod, error) {
	return CreateOrPatchPod(c, meta, transform)
}

func CreateOrPatchPod(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.Pod) *apiv1.Pod) (*apiv1.Pod, error) {
	cur, err := c.CoreV1().Pods(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		return c.CoreV1().Pods(meta.Namespace).Create(transform(&apiv1.Pod{ObjectMeta: meta}))
	} else if err != nil {
		return nil, err
	}
	return PatchPod(c, cur, transform)
}

func PatchPod(c clientset.Interface, cur *apiv1.Pod, transform func(*apiv1.Pod) *apiv1.Pod) (*apiv1.Pod, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	modJson, err := json.Marshal(transform(cur))
	if err != nil {
		return nil, err
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, apiv1.Pod{})
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, nil
	}
	glog.V(5).Infof("Patching Pod %s@%s.", cur.Name, cur.Namespace)
	return c.CoreV1().Pods(cur.Namespace).Patch(cur.Name, types.StrategicMergePatchType, patch)
}

func TryPatchPod(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.Pod) *apiv1.Pod) (result *apiv1.Pod, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().Pods(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = PatchPod(c, cur, transform)
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to patch Pod %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("Failed to patch Pod %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

func TryUpdatePod(c clientset.Interface, meta metav1.ObjectMeta, transform func(*apiv1.Pod) *apiv1.Pod) (result *apiv1.Pod, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.CoreV1().Pods(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.CoreV1().Pods(cur.Namespace).Update(transform(cur))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Pod %s@%s due to %v.", attempt, cur.Name, cur.Namespace, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("Failed to update Pod %s@%s after %d attempts due to %v", meta.Name, meta.Namespace, attempt, err)
	}
	return
}

// ref: https://github.com/coreos/prometheus-operator/blob/c79166fcff3dae7bb8bc1e6bddc81837c2d97c04/pkg/k8sutil/k8sutil.go#L64
// PodRunningAndReady returns whether a pod is running and each container has
// passed it's ready state.
func PodRunningAndReady(pod apiv1.Pod) (bool, error) {
	switch pod.Status.Phase {
	case apiv1.PodFailed, apiv1.PodSucceeded:
		return false, errors.New("pod completed")
	case apiv1.PodRunning:
		for _, cond := range pod.Status.Conditions {
			if cond.Type != apiv1.PodReady {
				continue
			}
			return cond.Status == apiv1.ConditionTrue, nil
		}
		return false, errors.New("pod ready condition not found")
	}
	return false, nil
}

func RestartPods(kubeClient clientset.Interface, namespace string, selector *metav1.LabelSelector) error {
	r, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return err
	}
	return kubeClient.CoreV1().Pods(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: r.String(),
	})
}

func GetString(m map[string]string, key string) string {
	if m == nil {
		return ""
	}
	return m[key]
}

func EnsureContainerDeleted(containers []apiv1.Container, name string) []apiv1.Container {
	for i, c := range containers {
		if c.Name == name {
			return append(containers[:i], containers[i+1:]...)
		}
	}
	return containers
}

func UpsertContainer(containers []apiv1.Container, nv apiv1.Container) []apiv1.Container {
	for i, vol := range containers {
		if vol.Name == nv.Name {
			containers[i] = nv
			return containers
		}
	}
	return append(containers, nv)
}

func UpsertVolume(volumes []apiv1.Volume, nv apiv1.Volume) []apiv1.Volume {
	for i, vol := range volumes {
		if vol.Name == nv.Name {
			volumes[i] = nv
			return volumes
		}
	}
	return append(volumes, nv)
}

func EnsureVolumeDeleted(volumes []apiv1.Volume, name string) []apiv1.Volume {
	for i, v := range volumes {
		if v.Name == name {
			return append(volumes[:i], volumes[i+1:]...)
		}
	}
	return volumes
}

func UpsertVolumeMount(mounts []apiv1.VolumeMount, nv apiv1.VolumeMount) []apiv1.VolumeMount {
	for i, vol := range mounts {
		if vol.Name == nv.Name {
			mounts[i] = nv
			return mounts
		}
	}
	return append(mounts, nv)
}

func EnsureVolumeMountDeleted(mounts []apiv1.VolumeMount, name string) []apiv1.VolumeMount {
	for i, v := range mounts {
		if v.Name == name {
			return append(mounts[:i], mounts[i+1:]...)
		}
	}
	return mounts
}

func UpsertEnvVar(vars []apiv1.EnvVar, nv apiv1.EnvVar) []apiv1.EnvVar {
	for i, vol := range vars {
		if vol.Name == nv.Name {
			vars[i] = nv
			return vars
		}
	}
	return append(vars, nv)
}

func EnsureEnvVarDeleted(vars []apiv1.EnvVar, name string) []apiv1.EnvVar {
	for i, v := range vars {
		if v.Name == name {
			return append(vars[:i], vars[i+1:]...)
		}
	}
	return vars
}

func AddFinalizer(pod *apiv1.Pod, finalizer string) {
	for _, name := range pod.Finalizers {
		if name == finalizer {
			return
		}
	}
	pod.Finalizers = append(pod.Finalizers, finalizer)
}

func HasFinalizer(pod *apiv1.Pod, finalizer string) bool {
	for _, name := range pod.Finalizers {
		if name == finalizer {
			return true
		}
	}
	return false
}

func RemoveFinalizer(pod *apiv1.Pod, finalizer string) {
	// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	r := pod.Finalizers[:0]
	for _, name := range pod.Finalizers {
		if name != finalizer {
			r = append(r, name)
		}
	}
	pod.Finalizers = r
}
