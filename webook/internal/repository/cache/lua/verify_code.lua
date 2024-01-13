local key = KEYS[1]
-- 使用次数，也就是验证次数
local cntKey = key..":cnt"
-- 预期中的验证码
local expectedCode = ARGV[1]
local code = redis.call("get", key)
-- 转成一个数字
local cnt = tonumber(redis.call("get", cntKey))
if cnt <= 0 then
    -- 说明用户一直输入错误 有人搞你
    -- 或者已经用过了，也是有搞你
    return -1
elseif expectedCode == code then
    -- 输入对了
    -- 用完，不能在用了
    redis.call("set", cntKey, -1)
    return 0
else
    -- 用户手一抖，输错了
    -- 可验证次数 -1
    redis.call("decr", cntKey, -1)
    return -2
end