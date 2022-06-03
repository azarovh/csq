package server

import "container/list"

type Item struct {
	Key   string
	Value string
}

type OrderedMap struct {
	l *list.List
	m map[string]*list.Element
}

func CreateOrderedMap() *OrderedMap {
	return &OrderedMap{l: list.New(), m: make(map[string]*list.Element)}
}

func (m *OrderedMap) Add(key string, value string) {
	item := Item{Key: key, Value: value}
	el := m.l.PushBack(&item)

	m.m[key] = el
}

func (m *OrderedMap) Remove(key string) {
	if el, found := m.m[key]; found {
		m.l.Remove(el)
		delete(m.m, key)
	}
}

func (m *OrderedMap) Get(key string) *string {
	if el, found := m.m[key]; found {
		return &el.Value.(*Item).Value
	}
	return nil
}

func (m *OrderedMap) GetAll() *list.List {
	return m.l
}
