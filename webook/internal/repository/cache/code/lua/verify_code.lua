local key = KEYS[1]
-- 用户输入的
local cntKey = key .. ":cnt"
-- 用户输入的验证码
local expectedCode = ARGV[1]

local cnt = tonumber(redis.call("get", cntKey))
local code = redis.call("get", key)

if cnt == nil or cnt <= 0 then
    -- 说明一直输入错误
    -- 或者已经用过了
    return -1
end

if expectedCode == code then
    -- 输入对了
    -- 用完，不能用了
    redis.call("set", cntKey, 0)
    return 0
else
    -- 用户手一抖，输入错误
    redis.call("decr", cntKey)
    return -2
end
