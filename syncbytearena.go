package util

import (
	"sync"
)

// SyncByteArena is concurrent safe version of ByteArena
type SyncByteArena struct {
	shards []byteArenaShard
}

type byteArenaShard struct {
	sync.Mutex
	ByteArena
}

const (
	defaultShards = 128
)

// NewSyncByteArena with defaultShards
func NewSyncByteArena(chunkAllocMinSize, chunkAllocMaxSize int) *SyncByteArena {
	return NewSyncByteArenaWithShards(chunkAllocMinSize, chunkAllocMaxSize, defaultShards)
}

// NewSyncByteArenaWithShards is ctor for SyncByteArena
func NewSyncByteArenaWithShards(chunkAllocMinSize, chunkAllocMaxSize, shardCount int) *SyncByteArena {
	shards := make([]byteArenaShard, 0, shardCount)
	for i := 0; i < shardCount; i++ {
		shards = append(shards, byteArenaShard{ByteArena: ByteArena{chunkAllocMinSize: chunkAllocMinSize, chunkAllocMaxSize: chunkAllocMaxSize}})
	}
	return &SyncByteArena{shards: shards}
}

// AllocBytes in the specified shard
func (sa *SyncByteArena) AllocBytes(shard, n int) []byte {
	idx := shard % len(sa.shards)

	return sa.shards[idx].AllocBytes(n)
}

func (shard *byteArenaShard) AllocBytes(n int) []byte {
	shard.Lock()
	defer shard.Unlock()

	return shard.ByteArena.AllocBytes(n)
}
