--你的验证码的 Redis 上的 key
-- phone_code:login:131xxxx1239
local key = KEYS[1]
-- 验证次数，我们一个验证码，最多重复三次
-- phone_code:login:131xxxx1239:cnt
local cntKey = key..":cnt"
-- 你的验证码 123456
local val = ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    -- key 存在，但是过期时间没有
    -- 系统错误 你的同事手贱
    return -2
elseif ttl == -2 or ttl < 540 then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    -- 完美，符合预期
    return 0
else
    -- 发送太频繁
    return -1
end