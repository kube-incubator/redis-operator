package failover

import (
	"errors"
	"time"

	redisv1alpha1 "github.com/kube-incubator/redis-operator/pkg/apis/redis/v1alpha1"
	k8s "github.com/kube-incubator/redis-operator/pkg/service/kubernetes"
	"github.com/kube-incubator/redis-operator/pkg/service/redis"
)

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
			if err := r.healer.MakeMaster(redisesIP[0]); err != nil {
				return err
			}
			break
		}
		minTime, err2 := r.checker.GetMinimumRedisPodTime(rf)
		if err2 != nil {
			return err2
		}
		if minTime > timeToPrepare {
			if err2 := r.healer.SetOldestAsMaster(rf); err2 != nil {
				return err2
			}
		} else {
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
	if err2 := r.checker.CheckAllSlavesFromMaster(master, rf); err2 != nil {
		if err3 := r.healer.SetMasterOnAll(master, rf); err3 != nil {
			return err3
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
