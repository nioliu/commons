package component

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

var mapLock = sync.Mutex{}      // 对map加锁
var locks = map[string]*confs{} // 对时间戳加锁
var clear = &task{
	Mutex: sync.Mutex{},
	doing: false,
}

type task struct {
	sync.Mutex
	doing bool
}

type confs struct {
	*sync.Mutex
	seq  int
	mark int
	t    string
}

// 0～11bit	12bits	序列号，用来对同一个毫秒之内产生不同的ID，可记录4095个
// 12～21bit	10bits	10bit用来记录机器ID，总共可以记录1024台机器
// 22～62bit	41bits	用来记录时间戳，这里可以记录69年
// 63bit	1bit	符号位，不做处理

// CreateSnowflakeId 三台机器 每台qps 10W -> 1ms 100w条
func CreateSnowflakeId(machineId string) string {
	if !clear.doing {
		clear.Mutex.Lock()
		go clearExpiredConf()
		clear.doing = true
		clear.Mutex.Unlock()
	}
	// 准备信息
	now := time.Now()
	milli := strconv.Itoa(int(now.UnixMilli()))
	key := milli + machineId

	// 同一毫秒加锁
	mapLock.Lock()
	conf, ok := locks[key]
	if !ok {
		conf = &confs{
			Mutex: &sync.Mutex{},
			seq:   0,
			mark:  0,
			t:     milli,
		}
		locks[key] = conf
	}
	conf.Lock()
	defer conf.Unlock()

	mapLock.Unlock()

	// 序列自增
	conf.seq += 1
	// 固定位数
	se := fmt.Sprintf("%.12d", conf.seq)
	// 生成唯一id
	return milli + machineId + se + strconv.Itoa(conf.mark)
}

// CreateShortSnowflakeId 短位生成
func CreateShortSnowflakeId(machineId string) string {
	if !clear.doing {
		clear.Mutex.Lock()
		go clearExpiredConf()
		clear.doing = true
		clear.Mutex.Unlock()
	}
	// 准备信息
	now := time.Now()
	//second := strconv.Itoa(int(now.Unix()))
	key := now.Format("20060102150405") + machineId

	// 同一秒加锁
	mapLock.Lock()
	conf, ok := locks[key]
	if !ok {
		conf = &confs{
			Mutex: &sync.Mutex{},
			seq:   0,
			mark:  0,
			t:     strconv.Itoa(int(now.Unix())),
		}
		locks[key] = conf
	}
	conf.Lock()
	defer conf.Unlock()

	mapLock.Unlock()

	// 序列自增
	conf.seq += 1
	// 固定位数
	se := fmt.Sprintf("%.3d", conf.seq)
	// 生成唯一id
	return key + se + strconv.Itoa(conf.mark)
}

// ClearExpiredConf clear expired conf from map
func clearExpiredConf() {
	// check again
	if clear.doing {
		return
	}
	for true {
		// clear map each 10 second
		time.Sleep(time.Second * 10)
		for k, v := range locks {
			if !v.TryLock() { // in use
				continue
			}
			unixTime, err := strconv.Atoi(v.t)
			if err != nil {
				continue
			}
			uti := time.UnixMilli(int64(unixTime))
			if time.Now().After(uti.Add(time.Second * 10)) {
				delete(locks, k)
			}
		}
	}
}
