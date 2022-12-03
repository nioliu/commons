package component

import (
	"sync"
	"testing"
	"time"
)

func TestCreateSnowflakeId(t *testing.T) {
	id := CreateSnowflakeId("123")
	println(id)
}

func TestOutF(t *testing.T) {
	conf := &confs{
		Mutex: &sync.Mutex{},
		seq:   0,
		mark:  0,
	}
	conf.Lock()
	defer conf.Unlock()
	go func() {
		//conf.Lock()
		println(conf)
		//conf.Unlock()
	}()
	println("123")
	time.Sleep(5 * time.Second)
}
