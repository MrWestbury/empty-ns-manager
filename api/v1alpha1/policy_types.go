package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EmptyNsPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec struct {
		NamespaceSelector struct {
			MatchLabels map[string]string `json:"matchLabels"`
		} `json:"namespaceSelector"`
		PolicyAction string `json:"policyAction"`
		Message      string `json:"message,omitempty"`
	} `json:"spec"`
}

// This model represents the EmptyNsPolicy custom resource used in the empty-ns-controller.
// It includes metadata and specification fields to define the policy's behavior.
