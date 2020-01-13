// poll
package poll

import (
	"container/list"
	"sync"
)

type Element struct {
	key     string
	element *list.Element
}

func (_self *Element) Chan() chan interface{} {
	return _self.element.Value.(chan interface{})
}

type MessagePoll struct {
	lock        sync.RWMutex
	chanListMap map[string]*list.List
}

func NewMessagePoll() *MessagePoll {
	return &MessagePoll{
		chanListMap: make(map[string]*list.List),
	}
}

func (_self *MessagePoll) Poll(key string) *Element {
	_self.lock.Lock()
	defer _self.lock.Unlock()

	pollChan := make(chan interface{}, 1)
	chanList, exist := _self.chanListMap[key]
	if !exist {
		chanList = list.New()
		_self.chanListMap[key] = chanList
	}
	return &Element{key: key, element: chanList.PushBack(pollChan)}
}

func (_self *MessagePoll) Contain(key string) bool {
	_self.lock.RLock()
	defer _self.lock.RUnlock()

	_, exist := _self.chanListMap[key]
	if exist {
		return true
	}
	return false
}

func (_self *MessagePoll) Keys() []string {
	_self.lock.RLock()
	defer _self.lock.RUnlock()

	keys := make([]string, 0, len(_self.chanListMap))
	for key := range _self.chanListMap {
		keys = append(keys, key)
	}
	return keys
}

func (_self *MessagePoll) Push(key string, data interface{}) bool {
	_self.lock.Lock()
	defer _self.lock.Unlock()

	chanList, exist := _self.chanListMap[key]
	if exist {
		for e := chanList.Front(); e != nil; e = e.Next() {
			pollChan := e.Value.(chan interface{})
			pollChan <- data
		}
		delete(_self.chanListMap, key)
		return true
	}
	return false
}

func (_self *MessagePoll) Remove(element *Element) bool {
	_self.lock.Lock()
	defer _self.lock.Unlock()

	chanList, exist := _self.chanListMap[element.key]
	if exist {
		chanList.Remove(element.element)
		return true
	}
	return false
}
