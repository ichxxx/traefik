package v1alpha1

import (
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDP is the CRD implementation of a Traefik UDP middleware.
// More info: https://doc.traefik.io/traefik/v3.0/middlewares/overview/
type MiddlewareUDP struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata"`

	Spec MiddlewareUDPSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true

// MiddlewareUDPSpec defines the desired state of a MiddlewareUDP.
type MiddlewareUDPSpec struct {
	// IPAllowList defines the IPAllowList middleware configuration.
	// This middleware accepts/refuses connections based on the client IP.
	// More info: https://doc.traefik.io/traefik/v3.0/middlewares/udp/ipallowlist/
	IPAllowList *dynamic.UDPIPAllowList `json:"ipAllowList,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MiddlewareUDPList is a collection of MiddlewareUDP resources.
type MiddlewareUDPList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata"`

	// Items is the list of MiddlewareUDP.
	Items []MiddlewareUDP `json:"items"`
}
