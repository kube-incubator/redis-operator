package v1alpha1

const (
	defaultRedisNumber          = 3
	defaultSentinelNumber       = 3
	defaultExporterImage        = "oliver006/redis_exporter"
	defaultExporterImageVersion = "v0.11.3"
	defaultRedisImage           = "redis:3.2-alpine"
	defaultSentinelImage        = "redis:3.2-alpine"
)

var (
	defaultSentinelCustomConfig = []string{"down-after-milliseconds 5000", "failover-timeout 10000"}
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

	if len(r.Spec.Redis.ExporterImage) == 0 {
		r.Spec.Redis.ExporterImage = defaultExporterImage
	}

	if len(r.Spec.Redis.ExporterVersion) == 0 {
		r.Spec.Redis.ExporterVersion = defaultExporterImageVersion
	}

	if len(r.Spec.Sentinel.CustomConfig) == 0 {
		r.Spec.Sentinel.CustomConfig = defaultSentinelCustomConfig
	}

}
