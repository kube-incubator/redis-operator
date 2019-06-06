package kubernetes

import (
	"k8s.io/client-go/kubernetes"
)

// Service is the K8s service entrypoint.
type Services interface {
	Deployment
	StatefulSet
}

type services struct {
	Deployment
	StatefulSet
}

// New returns a new Kubernetes service.
func New(kubeClient kubernetes.Interface) Services {
	return &services{
		Deployment:  NewDeploymentService(kubeClient),
		StatefulSet: NewStatefulSetService(kubeClient),
	}
}
