package failover

import (
	"errors"
	"sort"
	"strconv"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	redisscheme "github.com/kube-incubator/redis-operator/pkg/scheme/redis"
	k8s "github.com/kube-incubator/redis-operator/pkg/service/kubernetes"
	"github.com/kube-incubator/redis-operator/pkg/service/redis"
	"github.com/kube-incubator/redis-operator/pkg/util"
)

// RedisFailoverHeal defines the interface able to fix the problems on the redis cluster
type RedisFailoverHeal interface {
	MakeMaster(ip string) error
	SetOldestAsMaster(r *redisv1alpha1.Redis) error
	SetMasterOnAll(masterIP string, r *redisv1alpha1.Redis) error
	NewSentinelMonitor(ip string, monitor string, r *redisv1alpha1.Redis) error
	RestoreSentinel(ip string) error
	SetSentinelCustomConfig(ip string, r *redisv1alpha1.Redis) error
	SetRedisCustomConfig(ip string, r *redisv1alpha1.Redis) error
}

// RedisFailoverHealer is our implementation of RedisFailoverHeal interface
type RedisFailoverHealer struct {
	k8sService  k8s.Services
	redisClient redis.Client
}

// NewRedisFailoverHealer creates an object of the RedisFailoverHeal struct
func NewRedisFailoverHealer(k8sService k8s.Services, redisClient redis.Client) *RedisFailoverHealer {
	return &RedisFailoverHealer{
		k8sService:  k8sService,
		redisClient: redisClient,
	}
}

func (rfh *RedisFailoverHealer) MakeMaster(ip string) error {
	return rfh.redisClient.MakeMaster(ip)
}

// SetOldestAsMaster puts all redis to the same master, choosen by order of appearance
func (rfh *RedisFailoverHealer) SetOldestAsMaster(r *redisv1alpha1.Redis) error {
	ssp, err := rfh.k8sService.GetStatefulSetPods(r.Namespace, util.GetRedisName(r))
	if err != nil {
		return err
	}
	if len(ssp.Items) < 1 {
		return errors.New("number of redis pods are 0")
	}

	// Order the pods so we start by the oldest one
	sort.Slice(ssp.Items, func(i, j int) bool {
		return ssp.Items[i].CreationTimestamp.Before(&ssp.Items[j].CreationTimestamp)
	})

	newMasterIP := ""
	for _, pod := range ssp.Items {
		if newMasterIP == "" {
			newMasterIP = pod.Status.PodIP
			if err := rfh.redisClient.MakeMaster(newMasterIP); err != nil {
				return err
			}
		} else {
			if err := rfh.redisClient.MakeSlaveOf(pod.Status.PodIP, newMasterIP); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetMasterOnAll puts all redis nodes as a slave of a given master
func (rfh *RedisFailoverHealer) SetMasterOnAll(masterIP string, r *redisv1alpha1.Redis) error {
	ssp, err := rfh.k8sService.GetStatefulSetPods(r.Namespace, util.GetRedisName(r))
	if err != nil {
		return err
	}
	for _, pod := range ssp.Items {
		if pod.Status.PodIP == masterIP {
			if err := rfh.redisClient.MakeMaster(masterIP); err != nil {
				return err
			}
		} else {
			if err := rfh.redisClient.MakeSlaveOf(pod.Status.PodIP, masterIP); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewSentinelMonitor changes the master that Sentinel has to monitor
func (rfh *RedisFailoverHealer) NewSentinelMonitor(ip string, monitor string, r *redisv1alpha1.Redis) error {
	quorum := strconv.Itoa(int(redisscheme.GetQuorum(r)))
	return rfh.redisClient.MonitorRedis(ip, monitor, quorum)
}

// RestoreSentinel clear the number of sentinels on memory
func (rfh *RedisFailoverHealer) RestoreSentinel(ip string) error {
	return rfh.redisClient.ResetSentinel(ip)
}

// SetSentinelCustomConfig will call sentinel to set the configuration given in config
func (rfh *RedisFailoverHealer) SetSentinelCustomConfig(ip string, r *redisv1alpha1.Redis) error {
	return rfh.redisClient.SetCustomSentinelConfig(ip, r.Spec.Sentinel.CustomConfig)
}

// SetRedisCustomConfig will call redis to set the configuration given in config
func (rfh *RedisFailoverHealer) SetRedisCustomConfig(ip string, r *redisv1alpha1.Redis) error {
	return rfh.redisClient.SetCustomRedisConfig(ip, r.Spec.Redis.CustomConfig)
}
