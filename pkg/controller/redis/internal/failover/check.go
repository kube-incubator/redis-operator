package failover

import (
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	k8s "github.com/kube-incubator/redis-operator/pkg/service/kubernetes"
	"github.com/kube-incubator/redis-operator/pkg/service/redis"
	"github.com/kube-incubator/redis-operator/pkg/util"
)

// RedisFailoverCheck defines the interface able to check the correct status of a redis cluster
type RedisFailoverCheck interface {
	CheckAllSlavesFromMaster(master string, r *redisv1alpha1.Redis) error
	CheckSentinelNumberInMemory(sentinel string, r *redisv1alpha1.Redis) error
	CheckSentinelSlavesNumberInMemory(sentinel string, r *redisv1alpha1.Redis) error
	CheckSentinelMonitor(sentinel string, monitor string) error
	GetMasterIP(r *redisv1alpha1.Redis) (string, error)
	GetNumberMasters(r *redisv1alpha1.Redis) (int, error)
	GetRedisesIPs(r *redisv1alpha1.Redis) ([]string, error)
	GetSentinelsIPs(r *redisv1alpha1.Redis) ([]string, error)
	GetMinimumRedisPodTime(r *redisv1alpha1.Redis) (time.Duration, error)
}

// RedisFailoverChecker is our implementation of RedisFailoverCheck interface
type RedisFailoverChecker struct {
	k8sService  k8s.Services
	redisClient redis.Client
}

// NewRedisFailoverChecker creates an object of the RedisFailoverChecker struct
func NewRedisFailoverChecker(k8sService k8s.Services, redisClient redis.Client) *RedisFailoverChecker {
	return &RedisFailoverChecker{
		k8sService:  k8sService,
		redisClient: redisClient,
	}
}

// CheckAllSlavesFromMaster controlls that all slaves have the same master (the real one)
func (rfc *RedisFailoverChecker) CheckAllSlavesFromMaster(master string, r *redisv1alpha1.Redis) error {
	rips, err := rfc.GetRedisesIPs(r)
	if err != nil {
		return err
	}
	for _, rip := range rips {
		slave, err := rfc.redisClient.GetSlaveOf(rip)
		if err != nil {
			return err
		}
		if slave != "" && slave != master {
			return fmt.Errorf("slave %s don't have the master %s, has %s", rip, master, slave)
		}
	}
	return nil
}

// CheckSentinelNumberInMemory controls that the provided sentinel has only the living sentinels on its memory.
func (rfc *RedisFailoverChecker) CheckSentinelNumberInMemory(sentinel string, r *redisv1alpha1.Redis) error {
	nSentinels, err := rfc.redisClient.GetNumberSentinelsInMemory(sentinel)
	if err != nil {
		return err
	} else if nSentinels != r.Spec.Sentinel.Replicas {
		return errors.New("sentinels in memory mismatch")
	}
	return nil
}

// CheckSentinelSlavesNumberInMemory controls that the provided sentinel has only the expected slaves number.
func (rfc *RedisFailoverChecker) CheckSentinelSlavesNumberInMemory(sentinel string, r *redisv1alpha1.Redis) error {
	nSlaves, err := rfc.redisClient.GetNumberSentinelSlavesInMemory(sentinel)
	if err != nil {
		return err
	} else if nSlaves != r.Spec.Redis.Replicas-1 {
		return errors.New("redis slaves in sentinel memory mismatch")
	}
	return nil
}

// CheckSentinelMonitor controls if the sentinels are monitoring the expected master
func (rfc *RedisFailoverChecker) CheckSentinelMonitor(sentinel string, monitor string) error {
	actualMonitorIP, err := rfc.redisClient.GetSentinelMonitor(sentinel)
	if err != nil {
		return err
	}
	if actualMonitorIP != monitor {
		return errors.New("the monitor on the sentinel config does not match with the expected one")
	}
	return nil
}

// GetMasterIP connects to all redis and returns the master of the redis failover
func (rfc *RedisFailoverChecker) GetMasterIP(r *redisv1alpha1.Redis) (string, error) {
	rips, err := rfc.GetRedisesIPs(r)
	if err != nil {
		return "", err
	}
	masters := []string{}
	for _, rip := range rips {
		master, err := rfc.redisClient.IsMaster(rip)
		if err != nil {
			return "", err
		}
		if master {
			masters = append(masters, rip)
		}
	}

	if len(masters) != 1 {
		return "", errors.New("number of redis nodes known as master is different than 1")
	}
	return masters[0], nil
}

// GetNumberMasters returns the number of redis nodes that are working as a master
func (rfc *RedisFailoverChecker) GetNumberMasters(r *redisv1alpha1.Redis) (int, error) {
	nMasters := 0
	rips, err := rfc.GetRedisesIPs(r)
	if err != nil {
		return nMasters, err
	}
	for _, rip := range rips {
		master, err := rfc.redisClient.IsMaster(rip)
		if err != nil {
			return nMasters, err
		}
		if master {
			nMasters++
		}
	}
	return nMasters, nil
}

// GetRedisesIPs returns the IPs of the Redis nodes
func (rfc *RedisFailoverChecker) GetRedisesIPs(r *redisv1alpha1.Redis) ([]string, error) {
	redises := []string{}
	rps, err := rfc.k8sService.GetStatefulSetPods(r.Namespace, util.GetRedisName(r))
	if err != nil {
		return nil, err
	}
	for _, rp := range rps.Items {
		if rp.Status.Phase == corev1.PodRunning { // Only work with running pods
			redises = append(redises, rp.Status.PodIP)
		}
	}
	return redises, nil
}

// GetSentinelsIPs returns the IPs of the Sentinel nodes
func (rfc *RedisFailoverChecker) GetSentinelsIPs(r *redisv1alpha1.Redis) ([]string, error) {
	sentinels := []string{}
	rps, err := rfc.k8sService.GetDeploymentPods(r.Namespace, util.GetSentinelName(r))
	if err != nil {
		return nil, err
	}
	for _, sp := range rps.Items {
		if sp.Status.Phase == corev1.PodRunning { // Only work with running pods
			sentinels = append(sentinels, sp.Status.PodIP)
		}
	}
	return sentinels, nil
}

// GetMinimumRedisPodTime returns the minimum time a pod is alive
func (rfc *RedisFailoverChecker) GetMinimumRedisPodTime(r *redisv1alpha1.Redis) (time.Duration, error) {
	minTime := 100000 * time.Hour // More than ten years
	rps, err := rfc.k8sService.GetStatefulSetPods(r.Namespace, util.GetRedisName(r))
	if err != nil {
		return minTime, err
	}
	for _, redisNode := range rps.Items {
		if redisNode.Status.StartTime == nil {
			continue
		}
		start := redisNode.Status.StartTime.Round(time.Second)
		alive := time.Now().Sub(start)
		if alive < minTime {
			minTime = alive
		}
	}
	return minTime, nil
}
