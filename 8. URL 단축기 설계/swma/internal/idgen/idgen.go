package idgen

import (
	"fmt"
	"strconv"
	"os"
	"sync"
	"time"
	"unsafe"
)

const (
	timestampBits = 41
	datacenterBits = 5
	serverBits = 5
	sequenceBits = 12

	xEpoch = 1288834974657
)

type SnowflakeIdGenerator struct {
	mutex sync.Mutex
	timestamp int64
	datacenterId uint16
	serverId uint16
	sequence uint16
}

func clampBitMaxU16(value uint16, bits int) uint16 {
	return (value & ((uint16(1) << bits) - 1))
}

func clampBitMax64(value int64, bits int) int64 {
	return (value & ((int64(1) << bits) - 1))
}

func GetTimestampOfX() int64 {
	return time.Now().UnixMilli() - xEpoch
}

func NewGenerator(datacenterId uint16, serverId uint16) *SnowflakeIdGenerator {
	return &SnowflakeIdGenerator {
		datacenterId: datacenterId,
		serverId: serverId,
	}
}

func (g *SnowflakeIdGenerator) Update() {
	// Get current timestamp (from X epoch)
	now := GetTimestampOfX()

	if g.timestamp != now {
		g.sequence = 0
		g.timestamp = now
	} else {
		g.sequence = clampBitMaxU16(g.sequence + 1, sequenceBits)
		// Check for sequence conflicts
		if g.sequence == 0 {
			// Waiting for timestamp change
			for ; now == g.timestamp; now = GetTimestampOfX() {}
		}
	}
}

func (g *SnowflakeIdGenerator) Next() int64 {
	var id int64 = 0
	remain := int((unsafe.Sizeof(id) * 8) - 1)  // excludes sign bit(1)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Update generator
	g.Update()

	// Push (current timestamp)
	remain -= timestampBits
	id |= clampBitMax64(int64(g.timestamp), timestampBits) << remain

	// Push (Datacenter ID)
	remain -= datacenterBits
	id |= clampBitMax64(int64(g.datacenterId), datacenterBits) << remain

	// Push (Server ID)
	remain -= serverBits
	id |= clampBitMax64(int64(g.serverId), serverBits) << remain

	// Push (Sequence)
	remain -= sequenceBits
	id |= clampBitMax64(int64(g.sequence), sequenceBits) << remain

	return id
}

func main() {
	args := os.Args[1:]

	if len(args) < 4 {
		usage()
		return
	}

	delay, err := strconv.Atoi(args[0])
	if err != nil {
		usage()
		return
	}

	count, err := strconv.Atoi(args[1])
	if err != nil {
		usage()
		return
	}

	datacenterId, err := strconv.Atoi(args[2])
	if err != nil {
		usage()
		return
	}

	serverId, err := strconv.Atoi(args[3])
	if err != nil {
		usage()
		return
	}

	idGen := NewGenerator(uint16(datacenterId), uint16(serverId))

	for i := range count {
		fmt.Printf("New unique ID(%d/%d): %d\n", i + 1, count, idGen.Next())
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func usage() {
	fmt.Printf("Usage: %s [DELAY(ms)] [COUNT] [DATACENTER-ID] [SERVER-ID]\n", os.Args[0])
}
