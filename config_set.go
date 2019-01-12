package config

import (
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"strconv"
	"strings"
)

// Readonly disable set data to config.
// Usage:
//	config.LoadFiles(a, b, c)
//	config.Readonly()
func (c *Config) Readonly() {
	c.opts.Readonly = true
}

// Set a value by key string.
func (c *Config) Set(key string, val interface{}, setByPath ...bool) (err error) {
	// if is readonly
	if c.opts.Readonly {
		err = errors.New("the config instance in 'readonly' mode")
		return
	}

	// open lock
	c.lock.Lock()
	defer c.lock.Unlock()

	key = strings.Trim(strings.TrimSpace(key), ".")
	if key == "" {
		err = errors.New("the config key is cannot be empty")
		return
	}

	// is top key
	if !strings.Contains(key, ".") {
		c.data[key] = val
		return
	}

	// disable set by path.
	if len(setByPath) > 0 && !setByPath[0] {
		c.data[key] = val
		return
	}

	keys := strings.Split(key, ".")
	topK := keys[0]
	paths := keys[1:]

	var ok bool
	var item interface{}

	// find top item data based on top key
	if item, ok = c.data[topK]; !ok {
		// not found, is new add
		c.data[topK] = buildValueByPath(paths, val)
		return
	}

	switch typeData := item.(type) {
	case map[interface{}]interface{}: // from yaml
		dstItem := make(map[string]interface{})
		for k, v := range typeData {
			sk := fmt.Sprintf("%v", k)
			dstItem[sk] = v
		}

		// create a new item for the topK
		newItem := buildValueByPath(paths, val)
		// merge new item to old item
		err = mergo.Merge(&dstItem, newItem, mergo.WithOverride)
		if err != nil {
			return
		}

		c.data[topK] = dstItem
	case map[string]interface{}: // from json,toml
		// create a new item for the topK
		newItem := buildValueByPath(paths, val)
		// merge new item to old item
		err = mergo.Merge(&typeData, newItem, mergo.WithOverride)
		if err != nil {
			return
		}

		c.data[topK] = typeData
	case []interface{}: // is array
		index, err := strconv.Atoi(keys[1])
		if len(keys) == 2 && err == nil {
			if index <= len(typeData) {
				typeData[index] = val
			}

			c.data[topK] = typeData
		} else {
			err = errors.New("max allow 1 level for setting array value, current key: " + key)
			return err
		}
	default:
		err = errors.New("not supported value type, cannot setting value for the key: " + key)
	}
	return
}

/**
more setter: SetIntArr, SetIntMap, SetString, SetStringArr, SetStringMap
*/

// build new value by key paths
// "site.info" -> map[string]map[string]val
func buildValueByPath(paths []string, val interface{}) (newItem map[string]interface{}) {
	if len(paths) == 1 {
		return map[string]interface{}{paths[0]: val}
	}

	sliceReverse(paths)

	// multi nodes
	for _, p := range paths {
		if newItem == nil {
			newItem = map[string]interface{}{p: val}
		} else {
			newItem = map[string]interface{}{p: newItem}
		}
	}

	return
}

// reverse a slice. (slice 是引用，所以可以直接改变)
func sliceReverse(ss []string) {
	ln := len(ss)

	for i := 0; i < int(ln/2); i++ {
		li := ln - i - 1
		// fmt.Println(i, "<=>", li)
		ss[i], ss[li] = ss[li], ss[i]
	}
}
