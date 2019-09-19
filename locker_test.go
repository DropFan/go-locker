package locker_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	locker "github.com/DropFan/go-locker"
	"github.com/go-redis/redis"
)

func init() {
	initRedis()
}

func uninit() {
}

var rc redis.Cmdable

func GetRedis(ctx context.Context) redis.Cmdable {
	return rc
}

func initRedis() {
	rc = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	locker.SetRedis(context.TODO(), rc)
}

func getRandomStr(i int64) string {
	i += rand.Int63n(time.Now().UnixNano())
	s := strconv.FormatInt(i, 10) + "_" + time.Now().Format("20060102_030405.000000")
	return s
}

func Test_Lock_and_Unlock(t *testing.T) {
	defer func() {
		uninit()
	}()

	key := "test_locker_1"
	token1 := getRandomStr(time.Now().UnixNano())
	token2 := getRandomStr(time.Now().UnixNano())
	token3 := getRandomStr(time.Now().UnixNano())
	GetRedis(context.TODO()).Del(key)

	tests := []struct {
		name                 string
		key                  string
		value                string
		ctx                  context.Context
		expire               time.Duration
		waitExpire           bool
		waitLastExpire       bool
		lock                 bool
		unlock               bool
		wantLock             bool
		wantUnlock           bool
		unlockLastLocker     bool
		wantUnlockLastLocker bool
		doEvalFunc           locker.EvalFunc
	}{
		// TODO: Add test cases.
		{
			name:       "1-test_locker_1_basic",
			key:        key,
			value:      token1,
			expire:     time.Second * 10,
			lock:       true,
			unlock:     true,
			wantLock:   true,
			wantUnlock: true,
		},
		{
			name:       "2-test_locker_1_lock_not_unlock",
			key:        key,
			value:      token2,
			expire:     time.Second * 2,
			lock:       true,
			unlock:     false,
			wantLock:   true,
			wantUnlock: false,
		},
		{
			name:       "3-test_locker_1_lock_again",
			key:        key,
			value:      token2,
			expire:     time.Second * 2,
			waitExpire: true,
			lock:       true,
			unlock:     false,
			wantLock:   true,
			wantUnlock: false,
		},
		{
			name:   "4-test_locker_1_lock_after_expire",
			key:    key,
			value:  token1,
			expire: time.Second * 3,
			// waitExpire: true,
			lock:       true,
			unlock:     false,
			wantLock:   true,
			wantUnlock: false,
		},
		{
			name:       "5-test_locker_1_unlock_another_value",
			key:        key,
			value:      token3,
			expire:     time.Second * 3,
			waitExpire: false,
			lock:       false,
			unlock:     true,
			wantLock:   false,
			wantUnlock: false,
		},
		{
			name:   "6-test_locker_1_lock_again",
			key:    key,
			value:  token3,
			expire: time.Second * 2,
			// waitExpire: true,
			waitLastExpire: true,
			lock:           true,
			unlock:         false,
			wantLock:       true,
			wantUnlock:     false,
			// unlockLastLocker:     true,
			// wantUnlockLastLocker: true,
		},

		{
			name:                 "7-test_locker_1_lock_again_unlock_last_locker",
			key:                  key,
			value:                token2,
			expire:               time.Second * 2,
			waitExpire:           true,
			lock:                 false,
			unlock:               false,
			wantLock:             false,
			wantUnlock:           false,
			unlockLastLocker:     true,
			wantUnlockLastLocker: true,
		},

		{
			name:                 "8-test_locker_1_unlock_expired",
			key:                  key,
			value:                token1,
			expire:               time.Second * 1,
			lock:                 false,
			unlock:               true,
			wantLock:             false,
			wantUnlock:           false,
			unlockLastLocker:     true,
			wantUnlockLastLocker: false,
		},
		/*
		 */
	}
	// var lastLocker *locker.RedisLocker
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastT := tt
			if i > 1 {
				lastT = tests[i-1]
			}
			if tt.waitLastExpire {
				time.Sleep(lastT.expire + time.Microsecond*10)
			}

			if tt.ctx == nil {
				tt.ctx = context.Background()
			}

			m := locker.NewRedisLocker(tt.ctx, tt.key, tt.value, tt.expire)
			// m.SetLogFunc(logInfo)
			// m.SetLogErrorFunc(logError)
			if tt.doEvalFunc != nil {
				m.SetDoEvalFunc(tt.doEvalFunc)
			} else {
				m.SetDoEvalFunc(locker.GoRedisEvalFunc)
			}

			if tt.lock {
				if got := m.Lock(); got != tt.wantLock {
					t.Errorf("Lock() = %v, want %v", got, tt.wantLock)
				}
			}

			if tt.unlock {
				if got := m.Unlock(); got != tt.wantUnlock {
					t.Errorf("Unlock() = %v, want %v", got, tt.wantUnlock)
				}
			}

			if tt.unlockLastLocker {
				lastLocker := locker.NewRedisLocker(tt.ctx, lastT.key, lastT.value, lastT.expire)
				// lastLocker.SetDoEvalFunc(locker.GoRedisEvalFunc)
				// t.Logf("last val:%v, val:%v", lastLocker.GetValue(), m.GetValue())
				if got := lastLocker.Unlock(); got != tt.wantUnlockLastLocker {
					t.Errorf("lastLocker.Unlock() = %v, want %v", got, tt.wantUnlockLastLocker)
				}
			}

			if tt.waitExpire {
				time.Sleep(tt.expire + time.Microsecond*10)
			}
		})
	}
}

func Test_Info(t *testing.T) {
	ctx := context.TODO()

	t.Run("Test_Info", func(t *testing.T) {
		res := GetRedis(ctx).Info()
		t.Logf("info() result=\n%s", res)
	})
}
