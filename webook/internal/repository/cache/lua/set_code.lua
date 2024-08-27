-- 你的验证码在 Redis 上的 KEY
local key = KEYS[1]
-- 验证次数，我们一个验证码，最多重复三次，这个记录了验证了几次
local cntKey = key..":cnt"
local val = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))

if ttl == -1 then
--    key 存在，没有过期时间
    return -2
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    return 0
else
    return -1
end
