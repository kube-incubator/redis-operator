package util

import (
	"fmt"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/constants"
)

// GetRedisShutdownConfigMapName returns the name for redis configmap
func GetRedisShutdownConfigMapName(r *redisv1alpha1.Redis) string {
	if r.Spec.Redis.ShutdownConfigMap != "" {
		return r.Spec.Redis.ShutdownConfigMap
	}
	return GetRedisShutdownName(r)
}

// GetRedisName returns the name for redis resources
func GetRedisName(r *redisv1alpha1.Redis) string {
	return generateName(constants.RedisName, r.Name)
}

// GetRedisShutdownName returns the name for redis resources
func GetRedisShutdownName(r *redisv1alpha1.Redis) string {
	return generateName(constants.RedisShutdownName, r.Name)
}

// GetSentinelName returns the name for sentinel resources
func GetSentinelName(r *redisv1alpha1.Redis) string {
	return generateName(constants.SentinelName, r.Name)
}

func generateName(typeName, metaName string) string {
	return fmt.Sprintf("%s-%s-%s", constants.BaseName, typeName, metaName)
}
