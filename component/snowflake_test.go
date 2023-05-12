package component

import (
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func TestCreateSnowflakeId(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalln(err)
	}
	println(interfaces)

	inter := interfaces[0]

	s := sync.Map{}
	group := sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		group.Add(1)
		go func() {
			//id := CreateSnowflakeId(i.HardwareAddr.String())
			id2 := CreateShortSnowflakeId(inter.HardwareAddr.String())
			println(id2)
			_, ok := s.Load(id2)
			if ok {
				log.Fatalln("repeat", id2)
			} else {
				s.Store(id2, "")

			}
			group.Done()
		}()

	}
	group.Wait()

	//println(id)
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
