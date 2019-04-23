package sync

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/scheme/redis"
	"github.com/kube-incubator/redis-operator/pkg/staging/syncer"
)

// NewRedisConfigMapSyncer returns a new sync.Interface for reconciling Redis ConfigMap
func NewRedisConfigMapSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	cm := redis.GenerateRedisConfigMap(r, controllerLabels)
	return syncer.NewObjectSyncer("RedisConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}

// NewRedisShutdownConfigMapSyncer returns a new sync.Interface for reconciling Redis Shutdown ConfigMap
func NewRedisShutdownConfigMapSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	cm := redis.GenerateRedisShutdownConfigMap(r, controllerLabels)
	return syncer.NewObjectSyncer("RedisShutdownConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}

// NewSentinelConfigMapSyncer returns a new sync.Interface for reconciling Sentinel ConfigMap
func NewSentinelConfigMapSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	cm := redis.GenerateSentinelConfigMap(r, controllerLabels)
	return syncer.NewObjectSyncer("SentinelConfigMap", r, cm, c, scheme, func(existing runtime.Object) error {
		return nil
	})
}
