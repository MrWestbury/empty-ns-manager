package internal

import (
	"fmt"

	"github.com/mrwestbury/empty-ns-manager/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type PolicyWatcher struct {
	informer     cache.SharedIndexInformer
	PolicyAdd    func(policy v1alpha1.EmptyNsPolicy)
	PolicyDelete func(policyUid string)
	PolicyUpdate func(oldPolicy, newPolicy v1alpha1.EmptyNsPolicy)
	// Add fields if necessary
}

func NewPolicyWatcher(config *rest.Config) *PolicyWatcher {
	policyWatcher := &PolicyWatcher{}

	emtpyNsPolicy := schema.GroupVersionResource{Group: EMPTYNS_NAMESPACE, Version: "v1", Resource: "emptynspolicies"}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating dynamic client: ", err)
		return nil
	}

	factory := dynamicinformer.NewFilteredDynamicInformer(
		dynClient,
		emtpyNsPolicy,
		"",
		0,
		cache.Indexers{},
		nil,
	)

	policyWatcher.informer = factory.Informer()

	policyWatcher.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    policyWatcher.policyAdd,
		UpdateFunc: policyWatcher.policyUpdate,
		DeleteFunc: policyWatcher.policyDelete,
	})

	return policyWatcher
}

func (pw *PolicyWatcher) policyAdd(ob interface{}) {
	var policy v1alpha1.EmptyNsPolicy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(ob.(*unstructured.Unstructured).Object, &policy)
	if err != nil {
		fmt.Println("Error converting to EmptyNsPolicy:", err)
		return
	}
	if pw.PolicyAdd != nil {
		pw.PolicyAdd(policy)
	}
}

func (pw *PolicyWatcher) policyUpdate(oldObj, newObj interface{}) {
	var oldPolicy, newPolicy v1alpha1.EmptyNsPolicy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(oldObj.(*unstructured.Unstructured).Object, &oldPolicy)
	if err != nil {
		fmt.Println("Error converting old object to EmptyNsPolicy:", err)
		return
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(newObj.(*unstructured.Unstructured).Object, &newPolicy)
	if err != nil {
		fmt.Println("Error converting new object to EmptyNsPolicy:", err)
		return
	}
	if pw.PolicyUpdate != nil {
		pw.PolicyUpdate(oldPolicy, newPolicy)
	}
}

func (pw *PolicyWatcher) policyDelete(ob interface{}) {
	var policy v1alpha1.EmptyNsPolicy
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(ob.(*unstructured.Unstructured).Object, &policy)
	if err != nil {
		fmt.Println("Error converting to EmptyNsPolicy:", err)
		return
	}
	if pw.PolicyDelete != nil {
		pw.PolicyDelete(string(policy.UID))
	}
}

func (pw *PolicyWatcher) Run(stopCh <-chan struct{}) {
	pw.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, pw.informer.HasSynced) {
		fmt.Println("Timeout waiting for cache to sync")
		return
	}
}
