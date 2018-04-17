/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package consistenthash provides an implementation of a ring hash.
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// keys 中存放着虚拟节点 hash 后的值，并且会按照升序排序
// 虚拟节点可以在 hashMap 中找到实体节点，所有映射到虚拟节点的值都会被保存到实体节点的map 键值对中
// 查找方式使用的是二分查找算法
type Map struct {
	hash     Hash
	replicas int
	keys     []int // Sorted
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		//默认使用的函数是Hash方法是 crc32.ChecksumIEEE
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}

// Adds some keys to the hash.
// 添加一下hash 的key，虚拟节点数目就是replicas
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 对某个值生成副本数目个hash
			// 对于每个机器，如果是使用ip那么就会生成 “编号+ip" 的key，
			// 对于 192.168.0.1 就是 1192.168.0.1;2192.168.0.1;3192.168.0.1
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将hash值加入到key中
			m.keys = append(m.keys, hash)
			// 保存虚拟节点对应的实体节点
			m.hashMap[hash] = key
		}
	}
	// 升序排序
	sort.Ints(m.keys)
}

// Gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	// 查找满足条件的hash，使用的是一致性hash，所以是一个ring hash
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(m.keys) {
		idx = 0
	}

	return m.hashMap[m.keys[idx]]
}
