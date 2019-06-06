package kubernetes

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Deployment the Deployment service that knows how to interact with k8s to manage them
type Deployment interface {
	GetDeploymentPods(namespace, name string) (*corev1.PodList, error)
}

// DeploymentService is the service account service implementation using API calls to kubernetes.
type DeploymentService struct {
	kubeClient kubernetes.Interface
}

// NewDeploymentService returns a new Deployment KubeService.
func NewDeploymentService(kubeClient kubernetes.Interface) *DeploymentService {
	return &DeploymentService{
		kubeClient: kubeClient,
	}
}

// GetDeploymentPods will retrieve the pods managed by a given deployment
func (d *DeploymentService) GetDeploymentPods(namespace, name string) (*corev1.PodList, error) {
	deployment, err := d.kubeClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	labels := []string{}
	for k, v := range deployment.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return d.kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: selector})
}
