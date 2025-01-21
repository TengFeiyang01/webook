local zsetName = KEYS[1]  -- 有序集合的名称
local memberToIncrement = ARGV[1]   -- 要增加分数的元素

-- 获取指定元素的当前分数
local currentScore = redis.call("ZSCORE", zsetName, memberToIncrement)

if currentScore then
    -- 将分数加1
    local newScore = currentScore + 1
    -- 更新有序集合中指定元素的分数
    redis.call("ZADD", zsetName, newScore, memberToIncrement)
    return newScore
else
    -- 指定元素不存在
    return 0
end