package util

import (
	"github.com/kube-incubator/redis-operator/pkg/constants"
)

// MergeLabels merges all the label maps received as argument into a single new label map.
func MergeLabels(allLabels ...map[string]string) map[string]string {
	res := map[string]string{}

	for _, labels := range allLabels {
		if labels != nil {
			for k, v := range labels {
				res[k] = v
			}
		}
	}
	return res
}

// GetLabels returns the labels for the component with specific role
func GetLabels(component, role string) map[string]string {
	return generateLabels(component, role)
}

func generateLabels(component, role string) map[string]string {
	return map[string]string{
		"app":       constants.AppLabel,
		"component": component,
		component:   role,
	}
}
