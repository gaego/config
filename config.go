// Copyright 2012 GAEGo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package config provides persisted configuration maps that can be set
and retrieved using a unique key.

Each Config is stored in memory, memcache, and the datastore. This allows
Config settings to be shared between instances, while at the same time
elimating the cost associated with retrieving directly form the datastore.

Example Usage:

Set the config

	c := appengine.Context(r)
	appMap := map[string]string{
		"Title": "Storeski"
	}
	appConfig, err := config.GetOrInsert(c, "app", appMap)

Get the config in another package

	appConfig, err := config.Get(c, "app")
	appTitle := appConfig.Values"Title"] // "Storeski"

Edit and save the config

	appConfig.Values["Title"] = "Changed"
  err := appConfig.Put(c)
	appConfig2, err := config.Get(c, "app")
  appTitle := appConfig2.Values["Title"] // "Changed"

*/
package config

// TODO(kylefinley) move the JSON encoding of invalid datastore types
// to ds.

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"encoding/gob"
	"github.com/gaego/ds"
)

func init() {
	// Store Config entities in datastore, memcache and memory.
	ds.Register("Config", true, true, true)
}

// Config is the struct that is stored. The map[string]string{}
// is gob encoded before being passed to ds.
type Config struct {
	Key       *datastore.Key `datastore:"-"`
	ValuesGob []byte
	Values    map[string]string `datastore:"-"`
}

// SetKey set the Key property to a datastore.Key using the passed key
// as the StringID
func (cnfg *Config) SetKey(c appengine.Context, key string) {
	cnfg.Key = datastore.NewKey(c, "Config", key, 0, nil)
	return
}

// Encode is called prior to save. Any fields that need to be updated
// prior to save are updated here.
func (cnfg *Config) Encode() error {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(cnfg.Values)
	cnfg.ValuesGob = b.Bytes()
	return err
}

// Decode is called after the entity has been retrieved from the the ds.
func (cnfg *Config) Decode() error {
	b := bytes.NewBuffer(cnfg.ValuesGob)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&cnfg.Values)
	return err
}

// Put encodes the Values and saves the Config to the store.
func (cnfg *Config) Put(c appengine.Context) (err error) {
	if err = cnfg.Encode(); err != nil {
		return
	}
	key, err := ds.Put(c, cnfg.Key, cnfg)
	cnfg.Key = key
	return
}

// Get retieves a Config using the string key. If Config is not found
// it returns ds/error.ErrNoSuchEntity.
func Get(c appengine.Context, key string) (cnfg *Config, err error) {

	cnfg = new(Config)
	cnfg.SetKey(c, key)
	err = ds.Get(c, cnfg.Key, cnfg)
	if err != nil {
		return
	}
	err = cnfg.Decode()
	return
}

// GetOrInsert takes takes a key and map. If the key belongs to a previously
// saved Config it is returned; Otherwise the map is saved.
func GetOrInsert(c appengine.Context, key string,
	m map[string]string) (cnfg *Config, err error) {

	cnfg, err = Get(c, key)
	if err == nil {
		return
	}
	cnfg = new(Config)
	cnfg.Values = m
	cnfg.SetKey(c, key)
	err = cnfg.Put(c)
	return
}
