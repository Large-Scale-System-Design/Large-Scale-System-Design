package main

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

type Generator struct {
	mutex sync.Mutex
	timestamp int64
	datacenterId uint16
	serverId uint16
	sequence uint16
}

func NewGenerator(datacenterId uint16, serverId uint16) *Generator {
	return &Generator {
		datacenterId: datacenterId,
		serverId: serverId,
	}
}

func (g *Generator) Update() {
	// Get current timestamp (from X epoch)
	now := time.Now().UnixMilli() - xEpoch

	if g.timestamp != now {
		g.sequence = 0
		g.timestamp = now
	} else {
		g.sequence++
	}
}

func (g *Generator) Next() int64 {
	var id int64 = 0
	remain := int((unsafe.Sizeof(id) * 8) - 1)  // excludes sign bit(1)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Update generator
	g.Update()

	// Push (current timestamp)
	remain -= timestampBits
	id |= (int64(g.timestamp) & ((int64(1) << timestampBits) - 1)) << remain

	// Push (Datacenter ID)
	remain -= datacenterBits
	id |= (int64(g.datacenterId) & ((int64(1) << datacenterBits) - 1)) << remain

	// Push (Server ID)
	remain -= serverBits
	id |= (int64(g.serverId) & ((int64(1) << serverBits) - 1)) << remain

	// Push (Code)
	remain -= sequenceBits
	id |= (int64(g.sequence) & ((int64(1) << sequenceBits) - 1)) << remain

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
