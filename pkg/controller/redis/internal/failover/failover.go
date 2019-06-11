package failover

import (
	"errors"
	"time"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	k8s "github.com/kube-incubator/redis-operator/pkg/service/kubernetes"
	"github.com/kube-incubator/redis-operator/pkg/service/redis"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller_redis")

const (
	timeToPrepare = 2 * time.Minute
)

type RedisFailover struct {
	checker RedisFailoverCheck
	healer  RedisFailoverHeal
}

// NewRedisFailoverHandler returns a new RedisFailover
func NewRedisFailover(k8sService k8s.Services, redisClient redis.Client) *RedisFailover {
	checker := NewRedisFailoverChecker(k8sService, redisClient)
	healer := NewRedisFailoverHealer(k8sService, redisClient)
	return &RedisFailover{
		checker: checker,
		healer:  healer,
	}
}

func (r *RedisFailover) CheckAndHeal(rf *redisv1alpha1.Redis) error {
	fLogger := log.WithValues("Request.Namespace", rf.Namespace, "Request.Name", rf.Name)
	nMasters, err := r.checker.GetNumberMasters(rf)
	if err != nil {
		return err
	}
	switch nMasters {
	case 0:
		redisesIP, err := r.checker.GetRedisesIPs(rf)
		if err != nil {
			return err
		}
		if len(redisesIP) == 1 {
			fLogger.Info("Set redis master to " + redisesIP[0])
			if err := r.healer.MakeMaster(redisesIP[0]); err != nil {
				return err
			}
			break
		}
		minTime, err := r.checker.GetMinimumRedisPodTime(rf)
		if err != nil {
			return err
		}
		if minTime > timeToPrepare {
			fLogger.Info("Waiting more than 2 minutes, try to set the oldest node as master")
			if err := r.healer.SetOldestAsMaster(rf); err != nil {
				return err
			}
		} else {
			fLogger.Info("No master found, wait until failover")
			return nil
		}
	case 1:
		break
	default:
		return errors.New("More than one master, fix manually")
	}

	master, err := r.checker.GetMasterIP(rf)
	if err != nil {
		return err
	}

	if err := r.checker.CheckAllSlavesFromMaster(master, rf); err != nil {
		fLogger.Info("Set all slaves with the same master: " + master)
		if err := r.healer.SetMasterOnAll(master, rf); err != nil {
			return err
		}
	}

	redises, err := r.checker.GetRedisesIPs(rf)
	if err != nil {
		return err
	}
	for _, rip := range redises {
		if err := r.healer.SetRedisCustomConfig(rip, rf); err != nil {
			return err
		}
	}

	sentinels, err := r.checker.GetSentinelsIPs(rf)
	if err != nil {
		return err
	}
	for _, sip := range sentinels {
		if err := r.checker.CheckSentinelMonitor(sip, master); err != nil {
			fLogger.Info("Set sentinel to monitor " + master)
			if err := r.healer.NewSentinelMonitor(sip, master, rf); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.checker.CheckSentinelNumberInMemory(sip, rf); err != nil {
			if err := r.healer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.checker.CheckSentinelSlavesNumberInMemory(sip, rf); err != nil {
			if err := r.healer.RestoreSentinel(sip); err != nil {
				return err
			}
		}
	}
	for _, sip := range sentinels {
		if err := r.healer.SetSentinelCustomConfig(sip, rf); err != nil {
			return err
		}
	}
	return nil
}
