package locker

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// Locker .....
type Locker interface {
	// Lock ...
	Lock() bool
	// Refresh ...
	// Refresh() error
	// Unlock ...
	Unlock() bool
	// GetValue return value of locker, i.e token for unlock
	GetValue() string
}

// ensure RedisLocker implements the interface Locker
var _ Locker = new(RedisLocker)

// RedisLocker ...
type RedisLocker struct {
	key    string
	value  string
	ctx    context.Context
	expire time.Duration
	log    LogFunc
	logErr LogFunc
	// newCtx NewContextFunc
	doEval EvalFunc
}

// EvalFunc ...redis eval func
type EvalFunc func(ctx context.Context, script string, keys []string, values ...interface{}) (ret int, err error)

// NewRedisLocker ...
func NewRedisLocker(ctx context.Context, key string, val string, expire time.Duration) *RedisLocker {
	r := &RedisLocker{
		ctx:    ctx,
		key:    key,
		value:  val,
		expire: expire,

		log:    defaultLogfunc,
		logErr: defaultLogErrfunc,
		doEval: GoRedisEvalFunc,
	}

	return r
}

// GetValue ...
func (r *RedisLocker) GetValue() string {
	return r.value
}

func (r *RedisLocker) evalScript(id scriptID, keys []string, values ...interface{}) (ret int, err error) {

	if r.doEval == nil {
		return -1, errors.New("redis Eval func is not set")
	}

	script := id.Get()
	if script == "" {
		return -1, errors.New("internal error invalid scriptID")
	}

	defer func() {
		r.log(r.ctx, "_redisLocker.evalScript()||script=%s||key=%s||value=%s||err=%v||ret=%v", id, r.key, r.value, err, ret)
	}()

	var (
		ret0 interface{}
		ok   bool
	)

	ret0, err = r.doEval(r.ctx, script, keys, values...)

	// mv, err = GetRedis(r.ctx).Eval(lockScript, redis.KeyAndValue{r.key, r.value}, redis.KeyAndValue{r.key, expireStr})
	if err != nil {
		r.logErr(r.ctx, "_redisLocker.evalScript()||msg=eval failed||key=%v||err=%v", r.key, err)
		return -1, errors.New("doEval failed:" + err.Error())
	}

	ret, ok = ret0.(int)

	if !ok {
		r.logErr(r.ctx, "_redisLocker.evalScript()||msg=unexpected return||key=%v||value=%v", r.key, r.value)
		err = errors.New("doEval error:unexpected return")
		return -1, err
	}

	return
}

// Lock ...
func (r *RedisLocker) Lock() (locked bool) {

	var (
		err error
		ret int
		// ok        bool
		expireStr string
	)

	defer func() {
		r.log(r.ctx, "_redisLocker.Lock()||key=%s||value=%s||expire=%v||result=%v||err=%v||ret=%v", r.key, r.value, expireStr /* r.expire.Seconds() */, locked, err, ret)
	}()

	expireStr = strconv.FormatUint(uint64(r.expire.Seconds()), 10)
	ret, err = r.evalScript(lockScript, []string{r.key, r.key}, r.value, expireStr)
	if err != nil {
		r.logErr(r.ctx, "_redisLocker.Lock()||msg=lock failed||key=%v||value=%v||err=%v", r.key, r.value, err)
		return
	}
	switch ret {
	case 1:
		r.log(r.ctx, "_redisLocker.Lock()||msg=lock success||key=%v", r.key)
		locked = true
	case 2:
		r.log(r.ctx, "_redisLocker.Lock()||msg=relock with same token, expire success||key=%v", r.key)
		locked = true
	case 3:
		r.log(r.ctx, "_redisLocker.Lock()||msg=relock with same token, expire failed||key=%v", r.key)
	default:
		r.log(r.ctx, "_redisLocker.Lock()||msg=lock failed||key=%v||value=%v||ret=%v", r.key, r.value, ret)
	}

	return locked
}

// Unlock ...
func (r *RedisLocker) Unlock() (unlocked bool) {

	var (
		err     error
		removed int
	)

	defer func() {
		r.log(r.ctx, "_redisLocker.Unlock()||key=%s||value=%s||expire=%v||result=%v||err=%v||ret=%v", r.key, r.value, r.expire.Seconds(), unlocked, err, removed)
	}()

	removed, err = r.evalScript(unlockScript, []string{r.key}, r.value)

	if err != nil {
		r.logErr(r.ctx, "_redisLocker.Unlock()||msg=unlock failed||key=%v||value=%v||err=%v", r.key, r.value, err)
		return false
	}

	switch removed {
	case 0:
		r.log(r.ctx, "_redisLocker.Unlock()||msg=unlock success delete 0||key=%v", r.key)
		unlocked = true
	case 1:
		r.log(r.ctx, "_redisLocker.Unlock()||msg=unlock success delete 1||key=%v", r.key)
		unlocked = true
	case -1:
		fallthrough
	default:
		r.log(r.ctx, "_redisLocker.Unlock()||msg=unlock failed||key=%v||value=%v", r.key, r.value)
	}

	return unlocked
}
