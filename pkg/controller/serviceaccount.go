package controller

import (
	"fmt"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/pkg/api/v1"
)

func (vc *VaultController) runServiceAccountWatcher() {
	for vc.processNextServiceAccount() {
	}
}

func (vc *VaultController) processNextServiceAccount() bool {
	// Wait until there is a new item in the working queue
	key, quit := vc.saQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two serviceAccounts with the same key are never processed in
	// parallel.
	defer vc.saQueue.Done(key)

	// Invoke the method containing the business logic
	err := vc.syncSAToStdout(key.(string))
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		vc.saQueue.Forget(key)
		return true
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if vc.saQueue.NumRequeues(key) < vc.options.MaxNumRequeues {
		glog.Infof("Error syncing serviceAccount %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		vc.saQueue.AddRateLimited(key)
		return true
	}

	vc.saQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	glog.Infof("Dropping serviceAccount %q out of the queue: %v", key, err)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the serviceAccount to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (vc *VaultController) syncSAToStdout(key string) error {
	obj, exists, err := vc.saIndexer.GetByKey(key)
	if err != nil {
		glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a ServiceAccount, so that we will see a delete for one serviceAccount
		fmt.Printf("ServiceAccount %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a ServiceAccount was recreated with the same name
		fmt.Printf("Sync/Add/Update for ServiceAccount %s\n", obj.(*v1.ServiceAccount).GetName())
	}
	return nil
}
