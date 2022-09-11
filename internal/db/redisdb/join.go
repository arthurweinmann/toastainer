package redisdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

/*
	KEYS[1] = {toasterid}
	KEYS[2] = {toasterid}_exeinfo (concatenated exeid and tvsip)
	KEYS[3] = {exeid}
	ARGV[1] = maxjoiners
	ARGV[2] = exeinfo  (concatenated exeid and tvsip)
	ARGV[3] = tvsip
	ARGV[4] = joinable for seconds
	ARGV[5] = execution timeout in seconds
*/
var toggleNewExecutionScript = redis.NewScript(`
local maxjoiners = tonumber(ARGV[1])

if maxjoiners > 0 then
	if redis.call("EXISTS", KEYS[1]) == 1 then
		if redis.call("INCR", KEYS[1]) <= maxjoiners then
			-- return combined exeid and tvsip
			return redis.call("GET", KEYS[2])
		end

		-- reset counter for new execution
		redis.call("DEL", KEYS[1])
	end
else
	if redis.call("EXISTS", KEYS[2]) == 1 then
		-- return combined exeid and tvsip
		return redis.call("GET", KEYS[2])
	end
end

-- create new execution
redis.call("SET", KEYS[1], "1", "EX", ARGV[4])
redis.call("SET", KEYS[2], ARGV[2], "EX", ARGV[4])

-- create exeid reminder for forced request
redis.call("SET", KEYS[3], ARGV[3], "EX", ARGV[5])

return ARGV[2]
`)

func JoinOrCreateExecution(toasterID, exeid, tvsip string, maxConcurrentJoiners, joinableForSec, timeoutSec int) (string, string, error) {
	tmpExeid := exeid + "---" + tvsip

	res, err := toggleNewExecutionScript.Run(context.Background(), GetClient(), []string{toasterID, toasterID + "_exeinfo", exeid}, maxConcurrentJoiners, tmpExeid, tvsip, joinableForSec, timeoutSec).Text()
	if err != nil {
		return "", "", err
	}

	spl := strings.Split(res, "---")
	return spl[0], spl[1], nil
}

// Can be used for example after a code update to force new request to start a new updated toaster
func ForceRefreshAutojoin(toasterID string) error {
	return GetClient().Del(context.Background(), toasterID, toasterID+"_exeinfo").Err()
}

// Same as ForceRefreshAutojoin, but with a delay in case of an error not to overwhelm a TVS
func DelayedForceRefreshAutojoin(toasterID string, in time.Duration) error {
	pipe := GetClient().Pipeline()

	pipe.Expire(context.Background(), toasterID, in)
	pipe.Expire(context.Background(), toasterID+"_exeinfo", in)

	_, err := pipe.Exec(context.Background())
	return err
}

/*
	KEYS[1] = {exeid}
	ARGV[1] = tvsip
	ARGV[2] = execution timeout in seconds
*/
var joinForceExeIDScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
	return {redis.call("GET", KEYS[1]), 0}
end

-- create exeid reminder for forced request
redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])

return {ARGV[1], 1}
`)

func GetOrForceExeIDExecution(exeid, tvsip string, timeoutSec int) (string, bool, error) {
	res, err := joinForceExeIDScript.Run(context.Background(), GetClient(), []string{exeid}, tvsip, timeoutSec).Result()
	if err != nil {
		return "", false, err
	}

	tab, ok := res.([]interface{})
	if !ok {
		return "", false, fmt.Errorf("invalid redis response type: %T", res)
	}

	if len(tab) != 2 {
		return "", false, fmt.Errorf("invalid redis response array length: %x", len(tab))
	}

	tvsip, ok = tab[0].(string)
	if !ok {
		return "", false, fmt.Errorf("invalid redis response ip type: %T", tab[0])
	}

	isnew, ok := tab[1].(int64)
	if !ok {
		return "", false, fmt.Errorf("invalid redis response isnew type: %T", tab[1])
	}

	return tvsip, isnew == 1, nil
}
