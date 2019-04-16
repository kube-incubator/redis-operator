package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisSpec defines the desired state of Redis
// +k8s:openapi-gen=true
type RedisSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Redis defines the settings for redis cluster
	Redis RedisSettings `json:"redis,omitempty"`

	// Sentinel defines the settings for sentinel cluster
	Sentinel SentinelSettings `json:"sentinel,omitempty"`

	// HardAntiAffinity defines if the PodAntiAffinity on the Deployment and
	// Statefulset has to be hard (it's soft by default)
	HardAntiAffinity bool `json:"hardAntiAffinity,omitempty"`

	// NodeAffinity defines the rules for scheduling the Redis and Sentinel nodes
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`

	// SecurityContext defines which user and group the Sentinel and Redis containers run as
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	//Tolerations provides a way to schedule Pods on Tainted Nodes
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Replicas              int32          `json:"replicas,omitempty"`
	Resources             RedisResources `json:"resources,omitempty"`
	Exporter              bool           `json:"exporter,omitempty"`
	ExporterImage         string         `json:"exporterImage,omitempty"`
	ExporterVersion       string         `json:"exporterVersion,omitempty"`
	DisableExporterProbes bool           `json:"disableExporterProbes,omitempty"`
	Image                 string         `json:"image,omitempty"`
	Version               string         `json:"version,omitempty"`
	CustomConfig          []string       `json:"customConfig,omitempty"`
	Command               []string       `json:"command,omitempty"`
	ShutdownConfigMap     string         `json:"shutdownConfigMap,omitempty"`
	Storage               RedisStorage   `json:"storage,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Replicas     int32          `json:"replicas,omitempty"`
	Resources    RedisResources `json:"resources,omitempty"`
	CustomConfig []string       `json:"customConfig,omitempty"`
	Command      []string       `json:"command,omitempty"`
}

// RedisResources sets the limits and requests for a container
type RedisResources struct {
	Requests CPUAndMem `json:"requests,omitempty"`
	Limits   CPUAndMem `json:"limits,omitempty"`
}

// CPUAndMem defines how many cpu and ram the container will request/limit
type CPUAndMem struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

// RedisStorage defines the structure used to store the Redis Data
type RedisStorage struct {
	KeepAfterDeletion     bool                          `json:"keepAfterDeletion,omitempty"`
	EmptyDir              *corev1.EmptyDirVolumeSource  `json:"emptyDir,omitempty"`
	PersistentVolumeClaim *corev1.PersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
}

// RedisStatus defines the observed state of Redis
// +k8s:openapi-gen=true
type RedisStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Redis is the Schema for the redis API
// +k8s:openapi-gen=true
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
