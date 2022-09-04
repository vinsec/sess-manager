package providers

import (
	"container/list"
	"sync"
	"time"

	"github.com/vinsec/sess-manager/manager"
)

type MemProvider struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     *list.List
}

func (memProvider *MemProvider) SessionInit(sid string) (manager.Session, error) {
	memProvider.lock.Lock()
	defer memProvider.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{
		sid:          sid,
		timeAccessed: time.Now(),
		value:        v,
	}
	//GC 链表头插入 newsess
	element := memProvider.list.PushFront(newsess)
	memProvider.sessions[sid] = element
	return newsess, nil

}

func (memProvider *MemProvider) SessionRead(sid string) (manager.Session, error) {
	if element, ok := memProvider.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := memProvider.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

func (memProvider *MemProvider) SessionUpdate(sid string) error {
	memProvider.lock.Lock()
	defer memProvider.lock.Unlock()
	if element, ok := memProvider.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		memProvider.list.MoveToFront(element)
		return nil
	}
	return nil
}

func (memProvider *MemProvider) SessionDestroy(sid string) error {
	if element, ok := memProvider.sessions[sid]; ok {
		delete(memProvider.sessions, sid)
		memProvider.list.Remove(element)
		return nil
	}
	return nil
}

func (memProvider *MemProvider) SessionGC(maxLifeTime int64) {
	memProvider.lock.Lock()
	defer memProvider.lock.Unlock()

	for {
		element := memProvider.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxLifeTime) < time.Now().Unix() {
			memProvider.list.Remove(element)
			delete(memProvider.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}
