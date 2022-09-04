package providers

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/vinsec/sess-manager/manager"
)

type RedisProvider struct {
	lock sync.Mutex
	rdb  *redis.Client
	ctx  context.Context
}

func (rdsProvider *RedisProvider) SessionInit(sid string) (manager.Session, error) {
	rdsProvider.lock.Lock()
	defer rdsProvider.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{
		sid:          sid,
		timeAccessed: time.Now(),
		value:        v,
	}
	sessBytes, err := json.Marshal(newsess)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	err = rdsProvider.setSessFromRedis(sid, string(sessBytes))
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return newsess, nil
}

func (rdsProvider *RedisProvider) SessionRead(sid string) (manager.Session, error) {
	if sess, isExist := rdsProvider.getSessFromRedis(sid); isExist {
		return sess, nil
	} else {
		sess, err := rdsProvider.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

func (rdsProvider *RedisProvider) SessionUpdate(sid string) error {
	rdsProvider.lock.Lock()
	defer rdsProvider.lock.Unlock()
	if sess, isExist := rdsProvider.getSessFromRedis(sid); isExist {
		sess.timeAccessed = time.Now()
		sessBytes, err := json.Marshal(sess)
		if err != nil {
			log.Print(err)
			return err
		}
		err = rdsProvider.setSessFromRedis(sid, string(sessBytes))
		if err != nil {
			log.Print(err)
			return err
		}
		return nil
	}
	return nil
}

func (rdsProvider *RedisProvider) SessionDestroy(sid string) error {
	if _, isExist := rdsProvider.getSessFromRedis(sid); isExist {
		rdsProvider.deleteSessFromRedis(sid)
		return nil
	}
	return nil
}

func (rdsProvider *RedisProvider) SessionGC(maxLifeTime int64) {
	rdsProvider.lock.Lock()
	defer rdsProvider.lock.Unlock()

	res, err := rdsProvider.rdb.Keys(rdsProvider.ctx, "*").Result()
	if err != nil {
		log.Print(err)
		return
	}

	var sess SessionStore
	for _, sessStr := range res {
		_ = json.Unmarshal([]byte(sessStr), &sess)
		if (sess.timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
			rdsProvider.deleteSessFromRedis(sess.sid)
		}
	}

}

func (rdsProvider *RedisProvider) setSessFromRedis(sid, sessStr string) error {
	_, err := rdsProvider.rdb.Set(rdsProvider.ctx, sid, sessStr, 0).Result()
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}

func (rdsProvider *RedisProvider) getSessFromRedis(sid string) (*SessionStore, bool) {
	sessValue, err := rdsProvider.rdb.Get(rdsProvider.ctx, sid).Result()
	if err != nil {
		log.Print(err)
		return nil, false
	}

	var sess SessionStore
	err = json.Unmarshal([]byte(sessValue), &sess)
	if err != nil {
		log.Print(err)
		return nil, false
	}
	return &sess, true
}

func (rdsProvider *RedisProvider) deleteSessFromRedis(sid string) {
	_, err := rdsProvider.rdb.Del(rdsProvider.ctx, sid).Result()
	if err != nil {
		log.Print(err)
		return
	}
}
