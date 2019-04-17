package constants

const (
	BaseName                             = "base"
	AppLabel                             = "redis-operator"
	RedisName                            = "redis"
	SentinelName                         = "sentinel"
	RedisShutdownName                    = "redis-shutdown"
	RedisStorageVolumeName               = "redis-data"
	RedisConfigurationVolumeName         = "redis-config"
	RedisShutdownConfigurationVolumeName = "redis-shutdown-config"
	RedisRoleName                        = "redis"
	SentinelRoleName                     = "sentinel"
	RedisConfigFileName                  = "redis.conf"
	SentinelConfigFileName               = "sentinel.conf"
	HostnameTopologyKey                  = "kubernetes.io/hostname"

	GraceTime = 30
)

const (
	ExporterPort                 = 9121
	ExporterPortName             = "http-metrics"
	ExporterContainerName        = "redis-exporter"
	ExporterDefaultRequestCPU    = "25m"
	ExporterDefaultLimitCPU      = "50m"
	ExporterDefaultRequestMemory = "50Mi"
	ExporterDefaultLimitMemory   = "100Mi"
)
