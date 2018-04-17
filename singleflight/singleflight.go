/*
Copyright 2012 Google Inc.

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

// Package singleflight provides a duplicate function call suppression
// mechanism.
package singleflight

import "sync"

// call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.

// Do函数主要用于处理查询请求，并且合并重复的请求，如果在请求某一个指定key，而查询操作没有返回，这个时候如果紧接着
// 有请求相同key的，也就是g.m中已经有了这个key，那么就继续等待上次的查询，不会新建一个call，相反
// 如果没有这个call的记录就启动一个新的请求，请求处理玩之后会从m数据结构中删除这次的处理请求
// 如果某节点对某相同的key存在大量并发查询相，而该key的值不在缓存中，
// 这些并发查询就会触发大量的Load过程（从数据源或远端节点加载数据）。
// 但这些Load过程都是加载相同的数据，造成了资源的大量浪费。
// 为了避免这种现象，需要在查询触发Load过程前，先判断是否已经有相同的 Load过程正在运行。
// 如果存在，本次Load不执行，而是等正在运行的Load过程完成后直接使用其结果。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
