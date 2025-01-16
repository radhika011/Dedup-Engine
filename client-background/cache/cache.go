package cache

import (
	"client-background/common"
	"client-background/types"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
)

const CacheCapacity = (128 * 1024) / (32) // cachesize / hash size , eventually consts
// var CacheIsValid = true                   // eventually in env file

var Cache struct {
	Old              map[types.Hash]bool
	New              map[types.Hash]bool
	Size             int
	Mu               sync.Mutex
	OldHasDeletables bool
}

func Load() {

	Cache.Mu = sync.Mutex{}
	Cache.OldHasDeletables = true

	cacheFile, err := os.OpenFile(common.GetCacheFile(), os.O_RDONLY, 0644)
	if err != nil || !common.GetCacheFlag() {
		Cache.Old = make(map[types.Hash]bool)
	} else {
		dec := gob.NewDecoder(cacheFile)
		err = dec.Decode(&Cache.Old)
		if err != nil {
			fmt.Println("Error while decoding cache file")
			Cache.Old = make(map[types.Hash]bool)
		}
	}

	Cache.Size = len(Cache.Old)
	Cache.New = make(map[types.Hash]bool)
}

func Release() {
	Cache.Old = nil
	Cache.New = nil
	Cache.Size = 0
	Cache.OldHasDeletables = true
}

func Invalidate() {
	common.SetCacheFlag(false)
}

func Remove(hashes []types.Hash) {
	Cache.Mu.Lock()
	defer Cache.Mu.Unlock()
	for _, hash := range hashes {
		delete(Cache.New, hash)
	}
}

func Check(hash types.Hash) bool {

	Cache.Mu.Lock()
	defer Cache.Mu.Unlock()

	_, found := Cache.Old[hash]
	if found {
		delete(Cache.Old, hash)
		Cache.New[hash] = true
		return true
	}
	_, found = Cache.New[hash]
	if found {
		return true
	}

	if Cache.Size < CacheCapacity {
		Cache.New[hash] = true
		Cache.Size++
	} else if len(Cache.Old) == 0 {
		return false
	} else {

		for h := range Cache.Old {
			delete(Cache.Old, h)
			break
		}
		Cache.New[hash] = true
	}

	return false
}

func Persist() error {

	Cache.Mu.Lock()
	defer Cache.Mu.Unlock()

	Cache.Old = nil

	cacheFile, err := os.OpenFile(common.GetCacheFile(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	enc := gob.NewEncoder(cacheFile)
	err = enc.Encode(Cache.New)
	if err != nil {
		fmt.Println("Error while encoding cache file")
		return err
	}
	common.SetCacheFlag(true)
	// fmt.Println("Success")
	return nil

}
