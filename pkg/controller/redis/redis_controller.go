package redis

import (
	"context"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	"github.com/kube-incubator/redis-operator/pkg/controller/redis/internal/failover"
	"github.com/kube-incubator/redis-operator/pkg/controller/redis/internal/sync"
	k8s "github.com/kube-incubator/redis-operator/pkg/service/kubernetes"
	"github.com/kube-incubator/redis-operator/pkg/service/redis"
	"github.com/kube-incubator/redis-operator/pkg/staging/syncer"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_redis")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Redis Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	failover := failover.NewRedisFailover(k8s.New(kubernetes.NewForConfigOrDie(mgr.GetConfig())), redis.New())
	return &ReconcileRedis{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetRecorder("redis-controller"), failover: failover}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("redis-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Redis
	err = c.Watch(&source.Kind{Type: &redisv1alpha1.Redis{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to the resources that owned by the primary resource
	subresources := []runtime.Object{
		&corev1.Service{},
		&corev1.ConfigMap{},
		&appsv1.StatefulSet{},
		&appsv1.Deployment{},
	}

	for _, subresource := range subresources {
		err = c.Watch(&source.Kind{Type: subresource}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &redisv1alpha1.Redis{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRedis{}

// ReconcileRedis reconciles a Redis object
type ReconcileRedis struct {
	client   client.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	failover *failover.RedisFailover
}

// Reconcile reads that state of the cluster for a Redis object and makes changes based on the state read
// and what is in the Redis.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRedis) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Redis")

	// Fetch the Redis instance
	redis := &redisv1alpha1.Redis{}
	err := r.client.Get(context.TODO(), request.NamespacedName, redis)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.scheme.Default(redis)
	redis.SetDefaults()

	syncers := []syncer.Interface{
		sync.NewRedisServiceSyncer(redis, r.client, r.scheme),
		sync.NewSentinelServiceSyncer(redis, r.client, r.scheme),
		sync.NewRedisConfigMapSyncer(redis, r.client, r.scheme),
		sync.NewRedisShutdownConfigMapSyncer(redis, r.client, r.scheme),
		sync.NewSentinelConfigMapSyncer(redis, r.client, r.scheme),
		sync.NewSentinelDeploymentSyncer(redis, r.client, r.scheme),
		sync.NewRedisStatefulSetSyncer(redis, r.client, r.scheme),
	}

	if err = r.sync(syncers); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.failover.CheckAndHeal(redis); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileRedis) sync(syncers []syncer.Interface) error {
	for _, s := range syncers {
		if err := syncer.Sync(context.TODO(), s, r.recorder); err != nil {
			return err
		}
	}
	return nil
}
