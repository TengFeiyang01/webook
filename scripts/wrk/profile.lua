-- 初始token
token = nil
-- 需要请求验证的路由地址
path = "/users/login"
-- 第一次请求认证的请求方法
method = "POST"

-- 共有的headers设置
wrk.headers["Content-Type"] = "application/json"
wrk.headers["User-Agent"] = ""

-- 发送第一次authenticate认证请求
request = function ()
    body = '{"email": "123@qq.com","password": "Hello@#$world123"}'
    return wrk.format(method, path, wrk.headers, body)
end

response = function (status, headers, body)
    if not token and status == 200 then
        token = headers["X-Jwt-Token"]
        path = "/users/profile" -- 拿到token以后做资源地址的修改
        method = "GET" -- 请求profile需要GET方法
        wrk.headers["Authorization"] = string.format("Bear %s", token) -- 将获取到的token写入header中
    end
end