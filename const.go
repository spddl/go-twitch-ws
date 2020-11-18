package twitch

import "sync"

const joinRateQueueLimitTemplate = "_joinRateQueueLimit(%d) limit reached"
const authenticateRateQueueLimitTemplate = "_authenticateRateQueueLimit(%d) limit reached"
const queueRateLimitTemplate = "_queueRateLimit(%d) limit reached"
const queueRateLimitModOpTemplate = "_queueRateLimitModOp(%d) limit reached"
const queueRateLimitWhisperTemplate = "_queueRateLimitWhisper(%d) limit reached"

const debugTemplate = "> %s"

var ( // https://dev.twitch.tv/docs/irc/guide#command--message-limits
	// Authentication and join rate limits are:
	// 20 authenticate attempts per 10 seconds per user (200 for verified bots)
	authenticateRateLimitMessages         int      = 20
	verifiedauthenticateRateLimitMessages int      = 200
	authenticateRateLimitSeconds          int      = 10
	_authenticateRateQueueLimit           mutexInt = mutexInt{v: 0} // don't modify

	joinRateLimitMessages         int      = 20
	verifiedJoinRateLimitMessages int      = 200
	joinRateLimitSeconds          int      = 10
	_joinRateQueueLimit           mutexInt = mutexInt{v: 0} // don't modify

	// Command and message limits are:
	// Users sending commands or messages to channels in which they do not have Moderator or Operator status
	rateLimitMessages int      = 20
	rateLimitSeconds  int      = 30
	_queueRateLimit   mutexInt = mutexInt{v: 0} // don't modify

	// Users sending commands or messages to channels in which they have Moderator or Operator status
	rateLimitModOpMessages int      = 100
	rateLimitModOpSeconds  int      = 30
	_queueRateLimitModOp   mutexInt = mutexInt{v: 0} // don't modify

	// For Whispers, which are private chat message between two users:
	// Limit Applies to
	// 3 per second, up to 100 per minute
	// 40 accounts per day Users (not bots)
	rateLimitWhisperMinute          int      = 1
	rateLimitWhisperMinuteMessages  int      = 100
	rateLimitWhisperSeconds         int      = 1
	rateLimitWhisperSecondsMessages int      = 3
	_queueRateLimitWhisperMinute    mutexInt = mutexInt{v: 0} // don't modify
	_queueRateLimitWhisperSeconds   mutexInt = mutexInt{v: 0} // don't modify

	// 10 per second, up to 200 per minute
	// 500 accounts per day Known bots
	rateLimitKnownBotsWhisperMinute          int      = 1
	rateLimitKnownBotsWhisperMinuteMessages  int      = 200
	rateLimitKnownBotsWhisperSeconds         int      = 1
	rateLimitKnownBotsWhisperSecondsMessages int      = 10
	_queueRateLimitKnownBotsWhisperMinute    mutexInt = mutexInt{v: 0} // don't modify
	_queueRateLimitKnownBotsWhisperSeconds   mutexInt = mutexInt{v: 0} // don't modify

	// 20 per second, up to 1200 per minute
	// 100,000 accounts per day Verified bots
	rateLimitVerifiedBotsWhisperMinute          int      = 1
	rateLimitVerifiedBotsWhisperMinuteMessages  int      = 1200
	rateLimitVerifiedBotsWhisperSeconds         int      = 1
	rateLimitVerifiedBotsWhisperSecondsMessages int      = 20
	_queueRateLimitVerifiedBotsWhisperMinute    mutexInt = mutexInt{v: 0} // don't modify
	_queueRateLimitVerifiedBotsWhisperSeconds   mutexInt = mutexInt{v: 0} // don't modify
)

type mutexInt struct { // mutual-exclusion lock
	mutex sync.RWMutex
	v     int
}

func (i *mutexInt) get() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.v
}

func (i *mutexInt) add() {
	i.mutex.Lock()
	i.v += 1
	i.mutex.Unlock()
}
func (i *mutexInt) sub() {
	i.mutex.Lock()
	i.v -= 1
	i.mutex.Unlock()
}
