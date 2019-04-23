package sync

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/scheme/redis"
	"github.com/kube-incubator/redis-operator/pkg/staging/syncer"
)

// NewRedisStatefulSetSyncer returns a new sync.Interface for reconciling Redis StatefulSet
func NewRedisStatefulSetSyncer(r *redisv1alpha1.Redis, c client.Client, scheme *runtime.Scheme) syncer.Interface {
	statefulSet := redis.GenerateRedisStatefulSet(r, controllerLabels)
	return syncer.NewObjectSyncer("RedisStatefulSet", r, statefulSet, c, scheme, noFunc)
}
