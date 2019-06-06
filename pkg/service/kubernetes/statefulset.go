package kubernetes

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSet the StatefulSet service that knows how to interact with k8s to manage them
type StatefulSet interface {
	GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error)
	GetStatefulSetPods(namespace, name string) (*corev1.PodList, error)
}

// StatefulSetService is the service account service implementation using API calls to kubernetes.
type StatefulSetService struct {
	kubeClient kubernetes.Interface
}

// NewStatefulSetService returns a new StatefulSet KubeService.
func NewStatefulSetService(kubeClient kubernetes.Interface) *StatefulSetService {
	return &StatefulSetService{
		kubeClient: kubeClient,
	}
}

// GetStatefulSet will retrieve the requested statefulset based on namespace and name
func (s *StatefulSetService) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	statefulSet, err := s.kubeClient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSet, err
}

// GetStatefulSetPods will give a list of pods that are managed by the statefulset
func (s *StatefulSetService) GetStatefulSetPods(namespace, name string) (*corev1.PodList, error) {
	statefulSet, err := s.GetStatefulSet(namespace, name)
	if err != nil {
		return nil, err
	}
	labels := []string{}
	for k, v := range statefulSet.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return s.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
}
