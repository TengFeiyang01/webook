-- 具体业务
local key = KEYS[1]
-- 是阅读数，点赞数还是收藏数
local cntKey = ARGV[1]

local delta = tonumber(ARGV[2])

local exist=redis.call("EXISTS", key)
if exist == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    return 1
else
    return 0
end