package component

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

var mapLock = sync.Mutex{}      // 对map加锁
var locks = map[string]*confs{} // 对时间戳+machineId加锁

// 上一次最新的时间戳
type lastTimestamp struct {
	sync.Mutex       // 读写锁，更新的时候不能读，读的时候不能更新
	t          int64 // 时间戳数值
}

var lastT = lastTimestamp{Mutex: sync.Mutex{}}

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
	seq  int // 逻辑时钟，如果时间戳相等，那么就用内部的逻辑时钟
	mark int
	t    string
}

// 0～11bit	12bits	序列号，用来对同一个毫秒之内产生不同的ID，可记录4095个
// 12～21bit	10bits	10bit用来记录机器ID，总共可以记录1024台机器
// 22～62bit	41bits	用来记录时间戳，这里可以记录69年
// 63bit	1bit	符号位，不做处理

// CreateSnowflakeId 三台机器 每台qps 10W -> 1ms 100w条
func CreateSnowflakeId(machineId string) (string, error) {
	if !clear.doing {
		clear.Mutex.Lock()
		// 也是保证时间回拨的一个机制
		go clearExpiredConf()
		clear.doing = true
		clear.Mutex.Unlock()
	}
	// 准备信息
	now := time.Now()

	// 防止时钟回拨，只允许比上一个时间戳大的存在
	now, err := checkClockBack(now)
	if err != nil {
		return "", err
	}

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
	if conf.seq >= int(math.Pow10(12)) {
		// 休眠一秒，继续生成
		time.Sleep(time.Millisecond)
		return CreateShortSnowflakeId(machineId)
	}
	// 固定位数
	se := fmt.Sprintf("%.12d", conf.seq)
	// 生成唯一id
	return milli + machineId + se + strconv.Itoa(conf.mark), nil
}

// 检查是否发生了始终回拨
func checkClockBack(now time.Time) (time.Time, error) {
	n := 0
	lastT.Mutex.Lock()
	defer lastT.Mutex.Unlock()
	for {
		// 增加写锁，一旦被接受，就要更新last时间戳
		if now.UnixMilli() >= lastT.t {
			now = time.Now() // 更新时间戳
			lastT.t = now.UnixMilli()
			break
		} else if n >= 100 {
			return time.Time{}, fmt.Errorf("generate id failed, clock back happened")
		}
		n += 1
	}
	return now, nil
}

// CreateShortSnowflakeId 短位生成
func CreateShortSnowflakeId(machineId string) (string, error) {
	if !clear.doing {
		clear.Mutex.Lock()
		go clearExpiredConf()
		clear.doing = true
		clear.Mutex.Unlock()
	}
	// 准备信息
	now := time.Now()

	// 防止时钟回拨，只允许比上一个时间戳大的存在
	now, err := checkClockBack(now)
	if err != nil {
		return "", err
	}

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
	// 固定位数 这里是3位，也就是说同一时间戳内可以生成1000个id
	if conf.seq >= int(math.Pow10(3)) {
		// 休眠一秒，继续生成
		time.Sleep(time.Millisecond)
		return CreateShortSnowflakeId(machineId)
	}
	se := fmt.Sprintf("%.3d", conf.seq)
	// 生成唯一id
	return key + se + strconv.Itoa(conf.mark), nil
}

// ClearExpiredConf clear expired conf from map
func clearExpiredConf() {
	// check again
	if clear.doing {
		return
	}
	for {
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
			// 10s的时钟回拨？吓死人。
			if time.Now().After(uti.Add(time.Second * 10)) {
				delete(locks, k)
			}
		}
	}
}
