package util

import (
	"sync"
	"time"
)

// Snowflake 简易雪花算法ID生成器(单节点)
const (
	epoch        = 1735689600000 // 2025-01-01 毫秒
	nodeBits     = 5
	seqBits      = 12
	nodeMax      = -1 ^ (-1 << nodeBits)
	seqMask      = -1 ^ (-1 << seqBits)
)

type snowflake struct {
	mu        sync.Mutex
	timestamp int64
	node      int64
	seq       int64
}

var sf = &snowflake{node: 1}

// GenID 生成全局唯一 int64 ID
func GenID() int64 {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	now := time.Now().UnixMilli()
	if now < sf.timestamp {
		now = sf.timestamp
	}
	if now == sf.timestamp {
		sf.seq = (sf.seq + 1) & seqMask
		if sf.seq == 0 {
			now++
		}
	} else {
		sf.seq = 0
	}
	sf.timestamp = now
	return (now-epoch)<<(nodeBits+seqBits) | (sf.node << seqBits) | sf.seq
}