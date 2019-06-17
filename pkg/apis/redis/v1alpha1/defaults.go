package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultRedisNumber     = 3
	defaultSentinelNumber  = 3
	defaultRedisImage      = "redis:3.2-alpine"
	defaultSentinelImage   = "redis:3.2-alpine"
	defaultExporterImage   = "oliver006/redis_exporter:v0.11.3"
	defaultImagePullPolicy = corev1.PullPolicy("IfNotPresent")
)

var (
	defaultSentinelCustomConfig = []string{
		"down-after-milliseconds 5000",
		"failover-timeout 10000",
	}
)

// SetDefaults sets Redis field defaults
func (r *Redis) SetDefaults() {

	if r.Spec.Redis.Replicas == 0 {
		r.Spec.Redis.Replicas = defaultRedisNumber
	}

	if r.Spec.Sentinel.Replicas == 0 {
		r.Spec.Sentinel.Replicas = defaultSentinelNumber
	}

	if len(r.Spec.Redis.Image) == 0 {
		r.Spec.Redis.Image = defaultRedisImage
	}

	if len(r.Spec.Sentinel.Image) == 0 {
		r.Spec.Sentinel.Image = defaultSentinelImage
	}

	if len(r.Spec.Redis.ImagePullPolicy) == 0 {
		r.Spec.Redis.ImagePullPolicy = defaultImagePullPolicy
	}

	if len(r.Spec.Sentinel.ImagePullPolicy) == 0 {
		r.Spec.Sentinel.ImagePullPolicy = defaultImagePullPolicy
	}

	if len(r.Spec.Sentinel.CustomConfig) == 0 {
		r.Spec.Sentinel.CustomConfig = defaultSentinelCustomConfig
	}

	if len(r.Spec.Redis.Exporter.Image) == 0 {
		r.Spec.Redis.Exporter.Image = defaultExporterImage
	}

	if len(r.Spec.Redis.Exporter.ImagePullPolicy) == 0 {
		r.Spec.Redis.Exporter.ImagePullPolicy = defaultImagePullPolicy
	}
}
