package locker

type scriptID int

func (s scriptID) String() string {
	if s == lockScript {
		return "<lock script>"
	} else if s == unlockScript {
		return "<unlock script>"
	}
	return "<unknown script>"
}

// Get return script by ID
func (s scriptID) Get() (script string) {
	if s == lockScript {
		script = lockScriptStr
	} else if s == unlockScript {
		script = unlockScriptStr
	}

	return
}

const (
	lockScript   scriptID = 1
	unlockScript scriptID = 2

	lockScriptStr = `if redis.call('set',KEYS[1],ARGV[1],'EX',ARGV[2],'NX') then return 1 elseif redis.call('get',KEYS[1]) == ARGV[1] then if redis.call('expire',KEYS[1], ARGV[2]) then return 2 else return 3 end else return -1 end`

	unlockScriptStr = `local val = redis.call('get',KEYS[1]) if val == ARGV[1] then return redis.call('del',KEYS[1]) else return -1 end`
	/*
		lockScript = `
		if redis.call('set',KEYS[1],ARGV[1],'EX',ARGV[2],'NX') then
			return 1
		elseif redis.call('get',KEYS[1]) == ARGV[1] then
			if redis.call('expire',KEYS[1], ARGV[2]) then
				return 2
			else
				return 3
			end
		else
			return -1
		end
		`
		unlockScript = `
		local val = redis.call("get",KEYS[1])
		if val == ARGV[1] then
			return redis.call("del",KEYS[1])
		else
			return -1
		end
		`
	*/
)
