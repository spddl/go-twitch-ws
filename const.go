package twitch

import "sync"

var ( // https://dev.twitch.tv/docs/irc/guide#command--message-limits
	// Authentication and join rate limits are:
	// 20 authenticate attempts per 10 seconds per user (200 for verified bots)
	authenticateRateLimitMessages         int      = 20
	verifiedauthenticateRateLimitMessages int      = 200
	authenticateRateLimitSeconds          int      = 10
	_authenticateRateQueueLimit           MutexInt = MutexInt{v: 0} // don't modify

	joinRateLimitMessages         int      = 20
	verifiedJoinRateLimitMessages int      = 200
	joinRateLimitSeconds          int      = 10
	_joinRateQueueLimit           MutexInt = MutexInt{v: 0} // don't modify

	// Command and message limits are:
	// Users sending commands or messages to channels in which they do not have Moderator or Operator status
	rateLimitMessages int      = 20
	rateLimitSeconds  int      = 30
	_queueRateLimit   MutexInt = MutexInt{v: 0} // don't modify

	// Users sending commands or messages to channels in which they have Moderator or Operator status
	rateLimitModOpMessages int      = 100
	rateLimitModOpSeconds  int      = 30
	_queueRateLimitModOp   MutexInt = MutexInt{v: 0} // don't modify

	// For Whispers, which are private chat message between two users:
	// Limit Applies to
	// 3 per second, up to 100 per minute
	// 40 accounts per day Users (not bots)

	// 10 per second, up to 200 per minute
	// 500 accounts per day Known bots

	// 20 per second, up to 1200 per minute
	// 100,000 accounts per day Verified bots
)

type MutexInt struct { // mutual-exclusion lock
	Mutex sync.RWMutex
	v     int
}

func (i *MutexInt) Get() int {
	i.Mutex.RLock()
	defer i.Mutex.RUnlock()
	return i.v
}

func (i *MutexInt) Add() {
	i.Mutex.Lock()
	i.v += 1
	i.Mutex.Unlock()
}
func (i *MutexInt) Sub() {
	i.Mutex.Lock()
	i.v -= 1
	i.Mutex.Unlock()
}
