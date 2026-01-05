package internal

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type NamespaceManager struct {
	// Add fields if necessary
	discoveryClient *discovery.DiscoveryClient
	dyanmicClient   *dynamic.DynamicClient
	clientSet       *kubernetes.Clientset
}

func NewNamespaceManager(config *rest.Config) *NamespaceManager {
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating dynamic client: ", err)
		return nil
	}

	discClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		fmt.Println("Error creating discovery client: ", err)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating kubernetes clientset: ", err)
		return nil
	}

	return &NamespaceManager{
		discoveryClient: discClient,
		dyanmicClient:   dynClient,
		clientSet:       clientset,
	}
}

func (nm *NamespaceManager) ScanNamespaces(ctx context.Context) ([]v1.Namespace, error) {
	// Implement the logic to scan namespaces and apply policies
	nsList, err := nm.clientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error listing namespaces: ", err)
		return nil, err
	}

	namespaces := make([]v1.Namespace, 0, len(nsList.Items))
	namespaces = append(namespaces, nsList.Items...)
	return namespaces, nil
}

func (nm *NamespaceManager) isNSEmpty(namespace string) bool {
	// Implement logic to check if a namespace is empty
	resources, _ := nm.discoveryClient.ServerPreferredResources()

	for _, resourceGroup := range resources {
		if resourceGroup.GroupVersion == "v1" && resourceGroup.Kind == "Namespace" {
			continue
		}
		for _, resource := range resourceGroup.APIResources {
			// Skip subresources
			if resource.Name == "bindings" || resource.Name == "finalizers" || resource.Name == "status" {
				continue
			}
			// Check if the resource is namespaced
			if resource.Namespaced {

				resourceList, err := nm.dyanmicClient.Resource(schema.GroupVersionResource{
					Group:    resourceGroup.GroupVersion,
					Version:  resource.Version,
					Resource: resource.Name,
				}).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					continue
				}
				if len(resourceList.Items) > 0 {
					return false
				}
			}
		}
	}
	return true
}

func (nm *NamespaceManager) AddLabelToNamespace(ctx context.Context, namespace string, labels map[string]string) error {
	nsResource := nm.dyanmicClient.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	})

	ns, err := nsResource.Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	metadata, found, err := unstructured.NestedMap(ns.Object, "metadata")
	if err != nil || !found {
		return fmt.Errorf("metadata not found in namespace %s", namespace)
	}

	existingLabels, found, err := unstructured.NestedStringMap(metadata, "labels")
	if err != nil {
		return err
	}
	if !found {
		existingLabels = make(map[string]string)
	}

	for key, value := range labels {
		existingLabels[key] = value
	}

	if err := unstructured.SetNestedStringMap(metadata, existingLabels, "labels"); err != nil {
		return err
	}

	if err := unstructured.SetNestedMap(ns.Object, metadata, "metadata"); err != nil {
		return err
	}

	_, err = nsResource.Update(ctx, ns, metav1.UpdateOptions{})
	return err
}
