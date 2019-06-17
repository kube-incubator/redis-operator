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
}

// RedisSettings defines the specification of the redis cluster
type RedisSettings struct {
	Image             string                     `json:"image,omitempty"`
	ImagePullPolicy   corev1.PullPolicy          `json:"imagePullPolicy,omitempty"`
	Replicas          int32                      `json:"replicas,omitempty"`
	Resources         RedisResources             `json:"resources,omitempty"`
	CustomConfig      []string                   `json:"customConfig,omitempty"`
	Command           []string                   `json:"command,omitempty"`
	ShutdownConfigMap string                     `json:"shutdownConfigMap,omitempty"`
	Storage           RedisStorage               `json:"storage,omitempty"`
	Exporter          RedisExporter              `json:"exporter,omitempty"`
	Affinity          *corev1.Affinity           `json:"affinity,omitempty"`
	SecurityContext   *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	Tolerations       []corev1.Toleration        `json:"tolerations,omitempty"`
}

// SentinelSettings defines the specification of the sentinel cluster
type SentinelSettings struct {
	Image           string                     `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy          `json:"imagePullPolicy,omitempty"`
	Replicas        int32                      `json:"replicas,omitempty"`
	Resources       RedisResources             `json:"resources,omitempty"`
	CustomConfig    []string                   `json:"customConfig,omitempty"`
	Command         []string                   `json:"command,omitempty"`
	Affinity        *corev1.Affinity           `json:"affinity,omitempty"`
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	Tolerations     []corev1.Toleration        `json:"tolerations,omitempty"`
}

// RedisExporter defines the specification for the redis exporter
type RedisExporter struct {
	Enabled         bool              `json:"enabled,omitempty"`
	Image           string            `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
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
