package server

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrderedMap(t *testing.T) {
	m := CreateOrderedMap()
	assert.NotEqual(t, m, nil)
}

func TestAddItemOrderedMap(t *testing.T) {
	m := CreateOrderedMap()
	assert.NotEqual(t, m, nil)
	m.Add("a", "b")
	assert.Equal(t, *m.Get("a"), "b")
}

func TestRemoveItemOrderedMap(t *testing.T) {
	m := CreateOrderedMap()
	assert.NotEqual(t, m, nil)
	m.Add("a", "b")
	m.Remove("a")
	assert.Equal(t, m.GetAll().Len(), 0)
}

func TestRemoveItemNotEmptyOrderedMap(t *testing.T) {
	m := CreateOrderedMap()
	assert.NotEqual(t, m, nil)
	m.Add("a", "b")
	m.Add("b", "c")
	m.Remove("a")
	assert.Equal(t, m.GetAll().Len(), 1)
	assert.Equal(t, *m.Get("b"), "c")
}

func TestOrderOrderedMap(t *testing.T) {
	m := CreateOrderedMap()
	assert.NotEqual(t, m, nil)
	m.Add("a", "b")
	m.Add("b", "c")
	m.Add("c", "d")
	l := m.GetAll()
	el := l.Front()
	fmt.Println(el)
	assert.Equal(t, el.Value.(*Item).Key, "a")
	el = el.Next()
	assert.Equal(t, el.Value.(*Item).Key, "b")
	el = el.Next()
	assert.Equal(t, el.Value.(*Item).Key, "c")
}
