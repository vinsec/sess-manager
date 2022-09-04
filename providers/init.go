package providers

import (
	"container/list"
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/vinsec/sess-manager/manager"
)

// implement manager.Session
type SessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

var memProvider = &MemProvider{list: list.New()}
var redisProvider = &RedisProvider{ctx: context.Background()}

func init() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "10.24.3.65:8379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	redisProvider.rdb = rdb
	memProvider.sessions = make(map[string]*list.Element)

	manager.Register("redis", redisProvider)
	manager.Register("memory", memProvider)
}

func (st *SessionStore) Set(key, value interface{}) error {
	st.value[key] = value
	memProvider.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) Get(key interface{}) interface{} {
	memProvider.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	} else {
		return nil
	}
}

func (st *SessionStore) Delete(key interface{}) error {
	delete(st.value, key)
	memProvider.SessionUpdate(st.sid)
	return nil
}

func (st *SessionStore) SessionID() string {
	return st.sid
}
