package framework

func (f *Framework) DeletePod(name, namespace string) error {
	return f.KubeClient.CoreV1().Pods(namespace).Delete(name, deleteInBackground())
}
