# 用户基本功能与Gin|GORM 入门

## 定义用户基本接口

- 对于一个用户模块来说，最先要设计的接口就是：**注册和登录**
- 而后要考虑：**编辑和查看用户信息。**

即先定义 Web 接口，再去考虑后面的数据库设计之类的东西。

#### Handler 的用途

​	这里，我直接定义了一个 `UserHandler` ，之后**将所有和用户有关的路由都定义在了这个 `Handler` 上。**

同时也定义了一个 **`RegisterRoutes` 的方法，用来注册路由。**

```Go
// user.go
type UserHandler struct {
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	server.POST("/users/signup", u.SignUp)
	server.POST("/users/login", u.Login)
	server.POST("/users/edit", u.Edit)
	server.GET("/users/profile", u.Profile)
}
```

这里 `UserHandler` 上的 `RegisterRoutes` 是一种分散注册路由的做法，还有一种**集中式的做法，比如说在 `main` 函数里将所有的路由都注册好。**

- 集中式：
  - 优点：打开就能看到全部路由
  - 缺点：路由过多的时候，难以维护，查找起来不方便。
- 分散式：
  - 优点：比较有条理
  - 缺点：找路由的时候不方便

#### 用分组路由来简化注册

​	注意到我们所有的路由都有 `/users` 这个前缀，手抖写错一下可能就出问题了，这时候可以使用 **`Gin` 的分组路由功能**。

```Go
func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
    // 使用分组路由
	ug := server.Group("/users")

	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.Login)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
}
```

#### 目录结构

此时的目录结构如下图：

![image-20240103140237448](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240103140237448.png)

在 `webook` 顶级目录下有：

- `main` 文件，用于启动 `webook` 。
- 一个 `internal` 包，里面放着的就是我们所有的业务代码。
- 一个 `pkg` 包，**这是我们在整个项目中，沉淀出来的可以给别的项目使用的东西。**

等后续我们用到其他部分了，再继续增加别的目录。

#### 前端

前端代码是直接 `copy` 过来的，使用教程：

```shell
(base) PS E:\Go_Workspace\geektime-basic-go-master\master> cd .\webook-fe\
(base) PS E:\Go_Workspace\geektime-basic-go-master\master\webook-fe> npm run dev
```

出现这行就代表启动成功。

```shell
- ready started server on 0.0.0.0:3000, url: http://localhost:3000
```

如果出现 `'next' 不是内部或外部命令，也不是可运行的程序或批处理文件` 错误，尝试运行 `npm install`  后再去执行上述代码即可。

#### 注册页面

​	这时候我们需要考虑前端页面长成什么样，然后**根据前端页面的字段，来确定后端接口输入和输出是什么样子的。** **点击注册的时候，会发一个请求到后端** `/users/signup` **上，默认情况下，前端用的是 ** `JSON` ** 来传递数据。**

![image-20240103142858133](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240103142858133.png)

在点击注册后，我们观察请求标头中的荷载：

```json
{email: "123@qq.com", password: "123456", confirmPassword: "123456"}
```

这是是一个 `JSON` 形式的字符串，我们可以在后端的 `/users/signup` 中使用一个对应的结构体去接受前端发来的登录信息。

这里我们使用了**方法内部类**   `SignUpReq` 来接收数据。

- 优点：**除了这个** `SignUp` **方法能够使用** `SignUpReq` **其他方法都用不了**。

```go
type SignUpReq struct {
    // 反引号中的 email 代表这个字段在 json 中的名称是 email
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
var req SignUpReq
// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
// 解析错误，就会直接写回一个 400 的错误
if err := ctx.Bind(&req); err != nil {
	return
}
```

##### 后端处理

- **接受请求并校验**
- **调用业务逻辑处理请求**
- **根据业务逻辑处理结果返回响应**

###### 接收请求数据：`Bind` 方法

​	 `Bind` 方法是 `Gin` 里面最常用的用于接收请求的方法。`Bind` 方法会根据 `HTTP` 请求的 `Content-Type` 来决定怎么处理。比如我们的请求是 `JSON` 格式，`Content-Type` 是 `application/json`，那么 `Gin` 就会使用 `Json` 来反序列化。

###### **校验请求**

在我们这个注册的业务逻辑里面，校验分为两块：

- **邮箱要符合一定的格式**：也就是账号必须是一个合法的邮箱
- **密码和确认密码需要相等**：这是为了确保用户没有输错
- **密码要符合一定的规律**：要求用户输入的密码必须不少于八位，必须包含数字、特殊字符。

现在主要是通过二次验证这种机制来保证登录安全性。

使用**正则表达式**加密，校验的时候，只需要使用 `"github.com/dlclark/regexp2"` 中的 `MatchString` 方法就可以。

```go
emailRegexPattern    = "^[a-z0-9A-Z]+[- | a-z0-9A-Z . _]+@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-z]{2,}$"
emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
ok, err := emailExp.MatchString(req.Email)
```

##### 跨域问题

我们的请求是从 `localhost:3000` 这个前端发送到后端 `localhost:8090` 的

**这种就是跨域请求。协议、域名、和端口任意一个不同，都是跨域请求。**

正常来说，若不做额外处理，是没办法这样发请求的。

![image-20240104091836551](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240104091836551.png)

解决方法： **preflight请求** ：需要在 `preflight` 请求中告诉浏览器，**我这个 `localhost:8090` 能够接收 ** `localhost:3000` **过来的请求。**

**preflight请求** 的特征：`preflight` 请求会发到同一个地址上，使用 `Options` 方法，没有请求参数

![image-20240104093431844](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240104093431844.png)

![image-20240104093450681](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240104093450681.png)

###### 使用 `middleware` 来解决 `CORS`

`Gin` 里面也提供了一个 `middleware` 实现来解决跨域问题，在 `https://github.com/gin-gonic/contrib` ，执行 `go get github.com/gin-gonic/contrib` 之后就可以在代码中使用。

使用 `Gin` 中 `Engine` 上的 `Use` 方法来注册你的 `middleware` ，那么进到这个 `Engine` 中的所有请求，都会执行相应的代码。 接收 `*Context` 作为参数的方法就可以看作是 `HandlerFunc` 。

![image-20240104110942415](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240104110942415.png)

**跨域问题是因为发请求的协议+域名+端口和接收请求的协议+域名+端口对不上，比如说这里的 `localhost:3000` 发到 `localhost:8080` 上。**

## 用户注册：存储用户基本信息

我们使用 `docker-compose` 来搭建开发环境所需的依赖

![image-20240105092057045](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105092057045.png)

```yaml
version: '3.0'
services:
  mysql8:
    image: mysql:8.0.29
    restart: always
    command: --default-authentication-plugin=mysql_native_password
    environment:
      MYSQL_ROOT_PASSWORD: root
    volumes:
      #      设置初始化脚本
      - ./script/mysql/:/docker-entrypoint-initdb.d/
    ports:
      #      注意这里我映射为了 13316 端口
      - "13316:3306"
#  redis:
#    image: 'bitnami/redis:latest'
#    environment:
#      - ALLOW_EMPTY_PASSWORD=yes
#    ports:
#      - '6379:6379'
```

#### Docker Compose 基本命令

- `docker compose up` ：初始化 `docker compose` 并启动。
- `docker compose down` ：删除 `docker compose` 里面创建的各种容器。
- 若执行失败，没权限，执行以下命令即可
```shell
  sudo groupadd docker     #添加docker用户组
  sudo gpasswd -a $USER docker     #将登陆用户加入到docker用户组中
  newgrp docker     #更新用户组
  docker ps    #测试docker命令是否可以使用sudo正常使用
 ```

------------------------------------------------------------------------------------------------------

此时就需要考虑数据库相关的增删改查放哪里了？`UserHandler` ？**不可以，因为 `Handler` 只是负责和 `HTTP` 有关的东西，我们需要的是一个代表数据库抽象的东西。**

#### 引入 `Service-Repository-DAO` 三层结构

- **service：代表的是领域服务（domain service），代表一个业务的完整的处理过程。**
- **repository：代表领域对象的存储，也即存储数据的抽象**
- **dao：代表的是数据库的操作**

同时，我们还需要一个 domain，代表领域对象。

![image-20240105105347456](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105105347456.png)

`dao` 中的 `User` 模型：注意到，`dao` 里面操作的不是 `domain.User` ，而是新定义了一个类型。

这是因为：**`domain.User` 是业务概念，它不一定和数据库中表或者列完全对的上。而 `dao.User` 是直接映射到表里面的。**

那么问题就来了：**如何建表？**

![image-20240105112220350](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105112220350.png)

#### 密码加密

- 谁来加密？`service` 还是 `repository` 还是 `dao` ？
- 怎么加密？怎么选择一个安全的加密算法？

**PS：敏感信息应该是连日志都不能打**

###### 加密的位置：

![image-20240105113719864](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105113719864.png)

这里我们选择在 `service` 加密。

###### 如何加密

常见的加密算法（安全性逐步提高）：

1. `md5` 之类的哈希算法
2. 在 `1` 的基础上，引入了盐值(salt)，或者进行多次哈希。
3. `PBKDFF2` 、`BCrypt` 这一类随机盐值加密算法，同样的密文加密后的结果都不同。

这里我们使用 `BCrypt` 加密，`BCrypt` 加密后无法解密，只能同时比较加密后的值来确定两者是否相等。

优点：

- **不需要自己去生成盐值。**
- **不需要额外存储盐值。**
- **可以通过控制 `cost` 来控制加密性能。**
- **同样的文本，加密后的结果不同。**

使用：`golang.org/x/crypto`。

#### 怎么获得邮件冲突的错误？

答案就是，我们需要拿到数据库的唯一索引冲突错误。我们需要使用 `MySQL GO` 驱动的 `error` 定义，找到准确的错误。

具体而言，在 `dao` 这一层，我们转为了 `ErrUserDuplicateEmail` 错误，并且将这个错误一路网上返回。

![image-20240105135814571](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105135814571.png)

## 用户登录

#### 实现登录功能

![image-20240105144629664](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105144629664.png)

#### 登录校验

登陆成功之后，我要去 `/users/profile` 的时候， **我怎么知道用户登录没登陆**

无状态的 `HTTP` 协议：`HTTP` 并不会记录你的登录状态，因此需要记录一下这个状态，于是就有两个东西 `Cookie` 和 `Session` 。

`Cookie` 关键配置：

- `Domain` ：也就是 `Cookie` 可以用在什么域名下，按最小化原则来设定。
- `Path` ：`Cookie` 可以用在什么路径下，同样按最小化原则来设定。
- `Max-Age` 和 `Expires` ：过期时间，只保留必要时间。
- `Http-Only` ：设置为 `true` 的话，那么浏览器上的 `JS` 代码将无法使用这个 `Cookie` ，永远设置为 `true` 。
- `Secure`：只能用于 `HTTPS` 协议，生产环境永远设置为 `true` 。
- `SameSite` ：是否允许跨站发送 `Cookie` ，尽量避免。

`Session` ：关键数据我们希望放在后端，这个存储的东西就叫做 `Session` 。

![image-20240105162049868](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105162049868.png)

![image-20240105162319316](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105162319316.png)

#### 使用 `Gin` 的 `Session` 插件来实现登录功能

`https://github.com/gin-contrib/sessions` ![image-20240105163223432](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240105163223432.png)

```go
# 步骤一
store := cookie.NewStore([]byte("secret"))
server.Use(sessions.Sessions("mysession", store))
```

```go
# 步骤二
sess := sessions.Default(ctx)
// 我可以随便设置值了 放在 session 里的值
sess.Set("userId", user.Id)
sess.Save()
```

```go
func (l *LoginMiddleBuilder) IgnorePaths(path string) *LoginMiddleBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddleBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 不需要登录校验的
		if ctx.Request.URL.Path == "/users/login" || ctx.Request.URL.Path == "/users/signup" {
			return
		}

		sess := sessions.Default(ctx)
		id := sess.Get("UserId")
		if id != nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

# 步骤三
server.Use(middleware.NewLoginMiddleBuilder().
	IgnorePaths("/users/signup").
	IgnorePaths("/users/login").Build())
```

![image-20240107093421839](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240107093421839.png)

> 启动`mysql`坑点，可能自己电脑本身的 `mysql.exe`正在运行，且占用了 `3306`端口，此时通过 `docker compose up`启动时会报错。
>
> 解决方法：按下 `Win + R` 键，输入 `services.msc` 并按回车。在服务列表中找到`MySQL`服务，右键点击它，然后选择“停止”。或者通过命令行 `net stop mysql` 来停止 `MySQL`服务。

##### 使用`Redis`

![image-20240108084510912](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240108084510912.png)

```go
// 第一个参数是最大空闲连接数量
// 第二个就是 tmp, 不太可能用 udp
// 第三、四个就是连接信息和密码
// 第五第六就是两个 key
store, err := redis.NewStore(16,
	"tcp", "localhost:6379", "",
[]byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm"), 		 	  	[]byte("o6jdlo2cb9f9pb6h46fjmllw481ldebj"))

if err != nil {
	panic(err)
}
```

##### `Session` 参数

![image-20240108090839733](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240108090839733.png)

###### 通过 `session` 设置刷新时间。

```go
// web/user.go
sess.Options(sessions.Options{
    // 60 秒过期
    MaxAge: 60,
})

gob.Register(time.Time{})
// middleware/user.go
// 先拿到上一次的的更新时间
updateTime := sess.Get("update_time")
sess.Options(sessions.Options{
    MaxAge: 60,
})
now := time.Now().UnixMilli()
sess.Set("userId", id)
// 还没刷新过，刚登陆
if updateTime == nil {
    sess.Set("update_time", now)
    sess.Save()
    return
}
// updateTime 是有的
updateTimeVal, _ := updateTime.(int64)

// 超时未刷新
// now.Sub(updateTimeVal) > time.Minute 这是采用Time.time
if now-updateTimeVal > 60*1000 {
    sess.Set("update_time", now)
    sess.Save()
    return
}
```

> `gob.Register(time.Now())` 这里是用 Go 的方式编码解码 这个需要加上

## JWT（JSON Web Token）

![image-20240108110849206](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240108110849206.png)

### JWT简介

它主要由三部分组成：

- **Header：**头部，JWT的元数据，也就是描述这个`token`本身的数据，一个`JSON`对象。
- **Payload：**负载，数据内容，一个 `JSON` 对象。
- **Signature：**签名，根据 `header` 和 `token` 生成。

### JWT使用

```shell
go get github.com/golang-jwt/jwt/v5
```

在登录过程中， 使用 `JWT` 也是两步：

- **`JWT` 加密和解密数据。**
- **登录校验。**

#### JWT 改造跨域设置

我们的约定是，后端在 `x-jwt-token` 里面返回 `token`，前端在 `Authorization`里面带上 `token` 。

所以需要改造 `AllowHeaders` 和 `ExposeHeaders` 。

```go
server.Use(cors.New(cors.Config{
    AllowAllOrigins: false,
    AllowOrigins:    []string{"http://localhost:3000"},
    // 在使用 JWT 的时候，因为我们使用了 Authorizaition 的头部，所以需要加上
    AllowHeaders: []string{"Content-Type", "Authorization"},
    // 为了 JWT 这里的 Authorization 必须加上
    ExposeHeaders:    []string{"x-jwt-token", "Authorization"},
    AllowMethods:     []string{"POST", "GET", "PUT"},
    AllowCredentials: true,
    // 你不加这个 前端是拿不到的
    AllowOriginFunc: func(origin string) bool {
        if strings.HasPrefix(origin, "http://localhost") {
            return true
        }
        return strings.Contains(origin, "abc")
    },
    MaxAge: 12 * time.Hour,
}))
```

#### JWT登录校验

![image-20240108125653129](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240108125653129.png)

#### 接入JWT的步骤总结

- **要在 `Login` 接口中，登陆成功后生成 `JWT token`**
  - 在 `JWT token` 中写入数据。
  - 在 `JWT token` 通过 `HTTP Response Header x-jwt-token` 返回。
- **改造跨域中间件，允许前端访问 `x-jwt-token` 这个响应头。**

- **要接入 `JWT` 登录校验的 `Gin middleware` 。**
  - 读取 `JWT token` 。
  - 验证 `JWT token` 是否合法。
- **前端要携带 `JWT token`**

#### JWT的优缺点

和 `Session` 比起来，优点：

- 不依赖于第三方存储。
- 适合在分布式环境下使用。
- **提高性能**（因为没有 `Redis` 访问之类的）。

缺点：

- 对加密依赖非常大，比 `Session` 容易泄密。
- 最好**不要在 `JWT` 里面放置敏感信息。**

##  保护系统

### 限流

**我们的限流针对的对象是 `IP`，虽然 `IP` 不能实际上代表一个人，但这是我们比较好的选择了。**

![image-20240109105541451](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109105541451.png)

为什么使用 `Redis` 实现？

​	因为要**考虑到整个单体应用部署多个实例，用户的请求经过负载均衡之类的东西之后，就不一定落到同一个机器上了。** 因此需要使用 `Redis` 来计数。

### 安全问题

当前存在的问题： **一旦被攻击者拿到关键的 `JWT` 或者 `ssid` ，攻击者就能假冒你。**

方法：利用 `User-Agent` 增强安全性。

- `Login` 接口，在 `JWT token` 里面带上 `User-Agent` 信息。
- `JWT` 登录校验中间件，在里面比较 `User-Agent` 。

### 面试要点

![image-20240109113027029](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109113027029.png)

## kubernetes

`kubernetes` 是一个开源的`容器编排平台` ，简称 `k8s`。（管容器的）

### 基本概念

- **Pod**：实例。
- **Service**：逻辑上的服务，可以认为这是你业务上 `xxx` 服务的直接映射。
- **Deployment**：管理 `Pod` 的东西。

> 假如说你有一个 `Web` 应用，部署了三个实例，那么就是一个 `Web Service`，对应了三个 `Pod`。

#### `Docker` 启用 `k8s` 支持

在 `Docker` 里面开启 `Enable Kubernetes` 功能即可。

![image-20240109114350874](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109114350874.png)

#### 安装 `kubectl` 工具

打开 https://kubernetes.io/docs/tasks/tools/ 找到对应的平台，下载即可。

如果安装了 `curl`，使用如下命令即可：

```shell
curl.exe -LO "https://dl.k8s.io/release/v1.29.0/bin/windows/amd64/kubectl.exe"
```

![image-20240109115422481](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109115422481.png)

出现如上信息即为成功。

### 用 `Kubernetes` 部署 `Web` 服务器

去除依赖，由于我们的服务本身是依赖于 `MySQl` 和 `Redis` 的。所以我们需要先暂时去除这部分，再去部署。

#### 部署方案

​	我们的目标是**部署三个实例**，可以之间让用户访问。三个实例，这样即使一个崩溃了，也还有两个，比较不容易出问题。

<img src="C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109130501272.png" alt="image-20240109130501272" style="zoom: 33%;" />

![image-20240109130603713](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109130603713.png)

#### 准备 `Kubernetes` 容器镜像

- 首先在本地完成编译，生成一个可在 `Linux` 平台执行的 `webook` 可执行文件。（交叉编译）

  - 基本命令是：`GOOS=linux GOARCH=arm go build -o webook .`

  > window 是：
  >
  > ```shell
  > $env:GOOS="linux"
  > $env:GOARCH="arm"
  > go build -o webook .
  > ```

  ```dockerfile
  # 基础镜像
  FROM ubuntu:latest
  
  # 把编译后的打包进来这个镜像，放到工作目录 /app（你可以根据实际情况修改路径）
  WORKDIR /app/
  
  # 复制应用程序文件到容器中
  COPY webook /app/
  
  # CMD 是执行命令
  # 指定应用程序作为入口命令
  ENTRYPOINT ["/app/webook"]
  ```

- 其次是运行 `Docker` 命令，打包成一个镜像。

  - 基本命令是：`docker build -t xxx/webook:v0.0.1` ，其中`xxx`是自己起的名字。

  ```makefile
  .PHONY: docker
  docker:
  	# 把上次编译的东西删掉
  	@rm webook || true
  	@docker rmi -f ytf0609/webook:v0.0.1
  	# 指定编译成在 ARM 架构的 linux 操作系统上运行的可执行文件，
  	# 名字叫做 webook
  	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
  	# 这里你可以随便改这个标签，记得对应的 k8s 部署里面也要改
  	@docker build -t ytf0609/webook:v0.0.1 .
  ```

处于方便的目的，我打包成了一个 `make docker` 命令，**如果没有安装 `make` 工具，你可以一个个命令在命令行单独执行。**

##### 编写 `Deployment`

![image-20240109155544490](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109155544490.png)

> template 中的 app: webook 要和 metadata 左边 metadata 中的 name 对上

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
# 名称
  name: webook
# 规格说明
spec:
# 副本数量
  replicas: 3
  selector:
    matchLabels:
      app: webook
# template 描述的是 你的 POD 是什么样子的
  template:
    metadata:
      labels:
        # 按标签找
        app: webook
# POD 的具体信息
    spec:
      containers:
        - name: webook
          image: ytf/webook:v0.0.1
          ports:
            - containerPort: 8090
```

###### spec

![image-20240109155933542](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109155933542.png)

###### selector

![image-20240109160201853](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109160201853.png)

###### template

![image-20240109160433760](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109160433760.png)

###### image

`image` 就是镜像，显然我们这里我们使用的是 `Docker` 构建的镜像。

```yaml
spec:
  containers:
    - name: webook
	  image: ytf/webook:v0.0.1	
	  ports:
		- containerPort: 8090
```

##### 编写 `Service`

![image-20240109161042480](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240109161042480.png)

```yaml
apiVersion: v1
kind: Service
metadata:
# 代表我们的 webook 本体
  name: webook
spec:
# 规格说明，也即是相信说明这个服务是一个怎样的服务
  type: LoadBalancer
  selector:
    app: webook
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 8090
```

#### 启动服务

执行命令 `kubectl apply -f k8s-webook-service.yaml`，`apply`命令的意思是应用这个配置。

检查状态

```shell
(base) PS E:\Go_Workspace\geektime-basic-go-live\webook> kubectl apply -f k8s-webook-deployment.yaml
deployment.apps/webook created
(base) PS E:\Go_Workspace\geektime-basic-go-live\webook> kubectl get deployments
NAME     READY   UP-TO-DATE   AVAILABLE   AGE
(base) PS E:\Go_Workspace\geektime-basic-go-live\webook> kubectl get services     
NAME         TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
kubernetes   ClusterIP      10.96.0.1     <none>        443/TCP        4h25m
webook       LoadBalancer   10.97.2.153   localhost     80:32581/TCP   21s
```

### 使用 `Kubernetes` 部署 `MySQL`

`MySQl` 和前面的 `Go` 应用不同，他需要存储数据，也就是我们需要给他一个存储空间。

在 `k8s` 里面，**存储空间被抽象为 `PersistentVolume` （持久化卷）。**

![image-20240110090119102](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110090119102.png)

`MySQL Service` 和 `MySQL Deployment`

##### MySQL Service

```yaml
apiVersion: v1
kind: Service
metadata:
# 代表我们的 webook 本体
  name: webook
spec:
# 规格说明，也即是相信说明这个服务是一个怎样的服务
  type: LoadBalancer
  selector:
    app: webook
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 8090
```

##### MySQL Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webook-mysql
  labels:
    app: webook-mysql
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webook-mysql
  template:
    metadata:
      name: webook-mysql
      labels:
        app: webook-mysql
    spec:
      containers:
        - name: webook-mysql
          image: mysql:8.0
          env:
            - name: MYSQL_ROOT_PASSWORD
              value: root
          imagePullPolicy: IfNotPresent
          volumeMounts:
            # 这边要对应到 mysql 的数据存储的位置
            - mountPath: /var/lib/mysql
              name: mysql-storage
          ports:
            - containerPort: 3306
      restartPolicy: Always
      volumes:
        - name: mysql-storage
          persistentVolumeClaim:
            claimName: webook-mysql-claim-v4
```

![image-20240110095923347](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110095923347.png)

##### PersistentVlolumeClaim

一个容器需要什么存储资源，是通过 `PersistentVolumeClaim` 来声明的。

```yaml
 # pvc => PersistentVolumeClaim
# 开始描述 mysql 需要的存储资源的特征
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
# 这个是指我 MySQL 要用的东西
# 我 k8s 有什么
  name: webook-mysql-claim
  labels:
    app: webook-mysql-claim
spec:
# 这里的 manual 其实是一个我们自己写的，只是用来维护
  storageClassName: manual
# 访问模式，这里是控制能不能被多个 pod 读写
  accessModes:
    - ReadWriteOnce
# 究竟需要一些什么资源
  resources:
    requests:
# 需要一个 G 的容量
      storage: 1Gi
```

##### PersistentVolume

**持久化卷，表达我是一个什么样的存储结构。**

因此，**`PersistentVolume` 是存储本身说我有什么特性，而 `PersistentVolumeClaim` 是用的人说他需要什么特性。**

```yaml
# pvc => PersistentVolumeClaim
# 开始描述 mysql 需要的存储资源的特征
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
# 这个是指我 MySQL 要用的东西
# 我 k8s 有什么
  name: webook-mysql-claim-v4
spec:
# 这里的 manual 其实是一个我们自己写的，只是用来维护
  storageClassName: manualv4
# 访问模式，这里是控制能不能被多个 pod 读写
  accessModes:    
    - ReadWriteOnce
# 究竟需要一些什么资源
  resources:
    requests:
# 需要一个 G 的容量
      storage: 1Gi
```

```yaml
apiVersion: v1
# 这个指的是 我 k8s 有哪些 volume
kind: PersistentVolume
metadata:
  name: my-local-pv-v4
spec:
# 这个名称必须和 pvc 中的一致
  storageClassName: manualv4
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
#    - ReadOnlyMany
#    - ReadWriteMany
  hostPath:
    path: "/mnt/live"
```

##### 整体调用流程及其对应关系

![image-20240110110924769](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110110924769.png)

![image-20240110111542977](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110111542977.png)

![image-20240110111807109](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110111807109.png)

![image-20240110111955990](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110111955990.png)

>还有一个点是 测试的时候 填的数据库名应为 mysql 而非 webook

### 使用 `Kubernetes` 部署 `Redis`

部署最简单的单机 `Redis`。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: webook-redis
spec:
  selector:
    app: webook-redis
  ports:
    - protocol: TCP
      port: 11479
      # 外部访问的端口，必须是 30000-32767 之间
      nodePort: 30003
      # pod 暴露的端口
      targetPort: 6379
  type: NodePort
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webook-redis
  labels:
    app: webook-redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webook-redis
  template:
    metadata:
      name: webook-redis
      labels:
        app: webook-redis
    spec:
      containers:
        - name: webook-redis
          image: redis:latest
          imagePullPolicy: IfNotPresent
      restartPolicy: Always
```

其中 `port`、`nodePort`和 `targetPort`的含义

- `port`：是指 `Service`本身的，比如我们在 `Redis`里面连接信息用的就是 `demo-redis-service:6379`。
- `nodePort`：是指我在 `k8s`集群之外访问的端口，比如说我执行 `redis-cli -p 30379`。
- `targetPort`：是指 `Pod`上暴露的的端口。

![image-20240110114546532](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110114546532.png)

##### 测试 `Redis`

```shell
kubectl apply -f k8s-redis-deployment.yaml
kubectl apply -f k8s-redis-service.yaml
```

可以试试外部访问，直接使用 `redis-cli -h localhost -p 30003`就可以。（30003是你 `service`中的nodePort）

### Ingress 和 Ingress controller

一个 `Ingress controller` 可以控制住整个集群内部的所有 `Ingress` （符合条件的 `Ingress`）

或者这样说：

- `Ingress` 是你的配置
- `Ingress controller` 是执行这些配置的

#### 安装 `helm` 和 `ingress-nignx`

安装 `helm`：

```shell
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

安装 `ingress-nignx`

```shell
helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx
```

![image-20240110132140779](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110132140779.png)

运行 `kubectl apply -f k8s-ingress-nginx.yaml` ，然后去修改 `Hosts` 文件。

```shell
# localhost name resolution is handled within DNS itself.
#	127.0.0.1       localhost
#	::1             localhost
127.0.0.1 live.webook.com
```

完毕后去浏览器，输入 `live.webook.com/hello` 查看连接情况（此处需要关闭 `VPN`）（还需要更改 `hosts`文件）

## 集成 `Redis`、`MySQL` 启动。

![image-20240110143207123](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110143207123.png)

![image-20240110143224536](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240110143224536.png)

配置和 `main.go` 和 `config` 之后：

```shell
make docker
kubectl delete deployment webook
kubectl apply -f k8s-webook-deployment.yaml
# 此时再去查看 pod信息
kubectl get pods

ytf@DESKTOP-GVQFSJF:~/webook/webook$ kubectl get pods
NAME                            READY   STATUS    RESTARTS   AGE
webook-7fcdc64bd8-8bcqm         1/1     Running   0          11m
webook-7fcdc64bd8-hlqdp         1/1     Running   0          11m
webook-7fcdc64bd8-s9ctb         1/1     Running   0          11m
webook-mysql-6659b44c98-5mg6s   1/1     Running   0          16m
webook-redis-c889fd9b-lg6fq     1/1     Running   0          15m
```

> 坑点1：`postman`发送请求时显示 `socket hangup` 之类。
>
> ​	解决方法：`setting->proxy` 关闭代理即可。
>
> 坑点2：`kubectl get pods` 出现 `CrashLoopBackOff` 问题，查看日志（方法：`kubectl logs webook-7fcdc64bd8-8xsvj`，其中最后的是 `podID`）发现 `panic: Error 1049 (42000): Unknown database 'webook'`。
>
> ​	解决方法：`create database webook`即可。
>
> 坑点3：去浏览器运行 `live.webook.com` 时出错。
>
> ​	解决方法：关闭 `VPN`。

# 登录性能优化

## wrk安装

```shell
sudo apt install wrk
cd ~/wrk
make
// 加入环境变量
sudo mv wrk /usr/local/bin
```

### 压测前准备

- 启用 `JWT` 来测试。
- 修改 `/users/login` 对应的登录态保持时间。
- 去除 `ratelimit` 限制

#### 压测注册接口

在项目根目录下执行：

```shell
wrk -t1 -d1s -c2 -s ./scripts/wrk/signup.lua http://localhost:8090/users/signup
```

- `-t`：后面跟着的是线程数量
- `-d`：后面跟着的是持续时间
- -`c`：后面跟着的是并发数
- -`s`：后面跟着的是测试脚本

#### 压测登录接口

```shell
wrk -t1 -d1s -c2 -s ./scripts/wrk/login.lua http://localhost:8090/users/login
```

需要实现注册一个账号，然后修改 `login.lua` 中的相关信息

#### 压测 Profile 接口

```shell
wrk -t1 -d1s -c2 -s ./scripts/wrk/profile.lua http://localhost:8090/users/profile
```

需要修改 `User-Agent` 和 对应的 `Authorization`。

## 性能优化

前面的代码中，基本上性能瓶颈是出在两个地方：

- **加密算法**：耗费CPU，会令CPU成为瓶颈。
- **数据库查询**。

因此我们考虑引入 `Redis` 来优化性能，用户**会先从 `Redis` 里面查询**，而后在缓存未命中的情况下，才会直接从数据库中查询。

### 引入缓存

但是，我们并不会之间 `Redis`，而是引入一个缓存，来避免上层业务直接操作 `Redis`，同时我们也不是引入一个通用的 `Cache`，而是为业务编写专门的 `Cache`。

也就是 `UserCache`。

![image-20240112113901609](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240112113901609.png)

检测数据不存在的写法：

![image-20240112114212810](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240112114212810.png)

## Redis 数据结构

`Redis` 数据结构主要有：

- `string`：你存储的 `key` 是对应的值，是一个字符串。
- `list`：你存储的 `key` 对应的值，就是一个链表。
- `set`：你存储的 `key` 对应的值，是一个集合。
- `sorted set`：你存储的 `key` 对应的值，是一个有序集合。
- `hash`：你存储的 `key` 对应的值，是一个 `hash` 结构，也叫做字典结构、`map` 结构。

# 短信验证码登录

##  服务划分

- **一个独立的短信发送服务。**
- 在独立的短信发送服务基础上，**封装一个验证码功能。**
- 在验证码功能的基础上，**封装一个登录功能。**

这就是业务上的**超前半步设计**，也叫做**叠床架屋**。

![image-20240112133138357](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240112133138357.png)

![image-20240112142512058](C:\Users\ytf\AppData\Roaming\Typora\typora-user-images\image-20240112142512058.png)

