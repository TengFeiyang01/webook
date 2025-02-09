# 用户基本功能与Gin|GORM 入门

## 定义用户基本接口

- 对于一个用户模块来说，最先要设计的接口就是：**注册和登录**
- 而后要考虑：**编辑和查看用户信息。**

即先定义 Web 接口，再去考虑后面的数据库设计之类的东西。

### Handler 的用途

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

### 目录结构

此时的目录结构如下图：

![image-20240103140237448](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941098.png)

在 `webook` 顶级目录下有：

- `main` 文件，用于启动 `webook` 。
- 一个 `internal` 包，里面放着的就是我们所有的业务代码。
- 一个 `pkg` 包，**这是我们在整个项目中，沉淀出来的可以给别的项目使用的东西。**

等后续我们用到其他部分了，再继续增加别的目录。

### 前端

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

### 注册页面

​	这时候我们需要考虑前端页面长成什么样，然后**根据前端页面的字段，来确定后端接口输入和输出是什么样子的。** **点击注册的时候，会发一个请求到后端** `/users/signup` **上，默认情况下，前端用的是 ** `JSON` ** 来传递数据。**

![image-20240103142858133](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941343.png)

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

#### 后端处理

- **接受请求并校验**
- **调用业务逻辑处理请求**
- **根据业务逻辑处理结果返回响应**

##### 接收请求数据：`Bind` 方法

​	 `Bind` 方法是 `Gin` 里面最常用的用于接收请求的方法。`Bind` 方法会根据 `HTTP` 请求的 `Content-Type` 来决定怎么处理。比如我们的请求是 `JSON` 格式，`Content-Type` 是 `application/json`，那么 `Gin` 就会使用 `Json` 来反序列化。

##### **校验请求**

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

#### 跨域问题

我们的请求是从 `localhost:3000` 这个前端发送到后端 `localhost:8090` 的

**这种就是跨域请求。协议、域名、和端口任意一个不同，都是跨域请求。**

正常来说，若不做额外处理，是没办法这样发请求的。

![image-20240104091836551](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941617.png)

解决方法： **preflight请求** ：需要在 `preflight` 请求中告诉浏览器，**我这个 `localhost:8090` 能够接收 ** `localhost:3000` **过来的请求。**

**preflight请求** 的特征：`preflight` 请求会发到同一个地址上，使用 `Options` 方法，没有请求参数

![image-20240104093431844](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941477.png)

![image-20240104093450681](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941126.png)

##### 使用 `middleware` 来解决 `CORS`

`Gin` 里面也提供了一个 `middleware` 实现来解决跨域问题，在 `https://github.com/gin-gonic/contrib` ，执行 `go get github.com/gin-gonic/contrib` 之后就可以在代码中使用。

使用 `Gin` 中 `Engine` 上的 `Use` 方法来注册你的 `middleware` ，那么进到这个 `Engine` 中的所有请求，都会执行相应的代码。 接收 `*Context` 作为参数的方法就可以看作是 `HandlerFunc` 。

![image-20240104110942415](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941733.png)

**跨域问题是因为发请求的协议+域名+端口和接收请求的协议+域名+端口对不上，比如说这里的 `localhost:3000` 发到 `localhost:8080` 上。**

## 用户注册：存储用户基本信息

我们使用 `docker-compose` 来搭建开发环境所需的依赖

![image-20240105092057045](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941737.png)

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

### Docker Compose 基本命令

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

### 引入 `Service-Repository-DAO` 三层结构

- **service：代表的是领域服务（domain service），代表一个业务的完整的处理过程。**
- **repository：代表领域对象的存储，也即存储数据的抽象**
- **dao：代表的是数据库的操作**

同时，我们还需要一个 domain，代表领域对象。

![image-20240105105347456](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941966.png)

`dao` 中的 `User` 模型：注意到，`dao` 里面操作的不是 `domain.User` ，而是新定义了一个类型。

这是因为：**`domain.User` 是业务概念，它不一定和数据库中表或者列完全对的上。而 `dao.User` 是直接映射到表里面的。**

那么问题就来了：**如何建表？**

![image-20240105112220350](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941280.png)

### 密码加密

- 谁来加密？`service` 还是 `repository` 还是 `dao` ？
- 怎么加密？怎么选择一个安全的加密算法？

**PS：敏感信息应该是连日志都不能打**

##### 加密的位置：

![image-20240105113719864](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941959.png)

这里我们选择在 `service` 加密。

##### 如何加密

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

![image-20240105135814571](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941029.png)

## 用户登录

### 实现登录功能

![image-20240105144629664](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941586.png)

### 登录校验

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

![image-20240105162049868](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941252.png)

![image-20240105162319316](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941240.png)

### 使用 `Gin` 的 `Session` 插件来实现登录功能

`https://github.com/gin-contrib/sessions` ![image-20240105163223432](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941825.png)

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

![image-20240107093421839](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942560.png)

> 启动`mysql`坑点，可能自己电脑本身的 `mysql.exe`正在运行，且占用了 `3306`端口，此时通过 `docker compose up`启动时会报错。
>
> 解决方法：按下 `Win + R` 键，输入 `services.msc` 并按回车。在服务列表中找到`MySQL`服务，右键点击它，然后选择“停止”。或者通过命令行 `net stop mysql` 来停止 `MySQL`服务。

#### 使用`Redis`

![image-20240108084510912](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942330.png)

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

#### `Session` 参数

![image-20240108090839733](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942717.png)

##### 通过 `session` 设置刷新时间。

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

![image-20240108110849206](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942243.png)

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

![image-20240108125653129](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942023.png)

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

![image-20240109105541451](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942446.png)

为什么使用 `Redis` 实现？

​	因为要**考虑到整个单体应用部署多个实例，用户的请求经过负载均衡之类的东西之后，就不一定落到同一个机器上了。** 因此需要使用 `Redis` 来计数。

### 安全问题

当前存在的问题： **一旦被攻击者拿到关键的 `JWT` 或者 `ssid` ，攻击者就能假冒你。**

方法：利用 `User-Agent` 增强安全性。

- `Login` 接口，在 `JWT token` 里面带上 `User-Agent` 信息。
- `JWT` 登录校验中间件，在里面比较 `User-Agent` 。

### 面试要点

![image-20240109113027029](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942454.png)

## kubernetes

`kubernetes` 是一个开源的`容器编排平台` ，简称 `k8s`。（管容器的）

### 基本概念

- **Pod**：实例。
- **Service**：逻辑上的服务，可以认为这是你业务上 `xxx` 服务的直接映射。
- **Deployment**：管理 `Pod` 的东西。

> 假如说你有一个 `Web` 应用，部署了三个实例，那么就是一个 `Web Service`，对应了三个 `Pod`。

#### `Docker` 启用 `k8s` 支持

在 `Docker` 里面开启 `Enable Kubernetes` 功能即可。

![image-20240109114350874](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942560.png)

#### 安装 `kubectl` 工具

打开 https://kubernetes.io/docs/tasks/tools/ 找到对应的平台，下载即可。

如果安装了 `curl`，使用如下命令即可：

```shell
curl.exe -LO "https://dl.k8s.io/release/v1.29.0/bin/windows/amd64/kubectl.exe"
```

![image-20240109115422481](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942945.png)

出现如上信息即为成功。

### 用 `Kubernetes` 部署 `Web` 服务器

去除依赖，由于我们的服务本身是依赖于 `MySQl` 和 `Redis` 的。所以我们需要先暂时去除这部分，再去部署。

#### 部署方案

​	我们的目标是**部署三个实例**，可以之间让用户访问。三个实例，这样即使一个崩溃了，也还有两个，比较不容易出问题。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942070.png" alt="image-20240109130501272" style="zoom: 33%;" />

![image-20240109130603713](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942113.png)

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

> Deployment 配置：
>
> - replicas: 副本数,有多少个 pod
> - selector: 选择器
>   - matchLabels: 根据 label 选择哪些 pod 属于这个 deployment
>   - matchExpressions: 根据表达式选择哪些 pod 属于这个 deployment
> - template: 模板，定义 pod 的模板
>   - metadata: 元数据，定义 pod 的元数据
>   - spec: 规格，定义 pod 的规格
>     - containers: 容器，定义 pod 的容器
>       - name: 容器名称
>       - image: 容器镜像
>       - ports: 容器端口
>         - containerPort: 容器端口

![image-20240109155544490](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942870.png)

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

![image-20240109155933542](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943814.png)

###### selector

![image-20240109160201853](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943289.png)

###### template

![image-20240109160433760](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943881.png)

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

![image-20240109161042480](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943336.png)

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

![image-20240110090119102](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943986.png)

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

![image-20240110095923347](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943254.png)

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

![image-20240110110924769](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943570.png)

![image-20240110111542977](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943291.png)

![image-20240110111807109](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943291.png)

![image-20240110111955990](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943171.png)

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

![image-20240110114546532](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943786.png)

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

![image-20240110132140779](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943548.png)

运行 `kubectl apply -f k8s-ingress-nginx.yaml` ，然后去修改 `Hosts` 文件。

```shell
# localhost name resolution is handled within DNS itself.
#	127.0.0.1       localhost
#	::1             localhost
127.0.0.1 live.webook.com
```

完毕后去浏览器，输入 `live.webook.com/hello` 查看连接情况（此处需要关闭 `VPN`）（还需要更改 `hosts`文件）

## 集成 `Redis`、`MySQL` 启动。

![image-20240110143207123](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943757.png)

![image-20240110143224536](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180943177.png)

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
sudo apt install script
cd ~/script
make
// 加入环境变量
sudo mv script /usr/local/bin
```

### 压测前准备

- 启用 `JWT` 来测试。
- 修改 `/users/login` 对应的登录态保持时间。
- 去除 `ratelimit` 限制

#### 压测注册接口

在项目根目录下执行：

```shell
script -t1 -d1s -c2 -s ./scripts/script/signup.lua http://localhost:8090/users/signup
```

- `-t`：后面跟着的是线程数量
- `-d`：后面跟着的是持续时间
- -`c`：后面跟着的是并发数
- -`s`：后面跟着的是测试脚本

#### 压测登录接口

```shell
script -t1 -d1s -c2 -s ./scripts/script/login.lua http://localhost:8090/users/login
```

需要实现注册一个账号，然后修改 `login.lua` 中的相关信息

#### 压测 Profile 接口

```shell
script -t1 -d1s -c2 -s ./scripts/script/profile.lua http://localhost:8090/users/profile
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

![image-20240112113901609](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180944335.png)

检测数据不存在的写法：

![image-20240112114212810](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180944252.png)

## Redis 数据结构

`Redis` 数据结构主要有：

- `string`：你存储的 `key` 是对应的值，是一个字符串。
- `list`：你存储的 `key` 对应的值，就是一个链表。
- `set`：你存储的 `key` 对应的值，是一个集合。
- `sorted set`：你存储的 `key` 对应的值，是一个有序集合。
- `hash`：你存储的 `key` 对应的值，是一个 `hash` 结构，也叫做字典结构、`map` 结构。

# 短信验证码登录、面向接口编程和依赖注入

## 服务划分

- **一个独立的短信发送服务。**
- 在独立的短信发送服务基础上，**封装一个验证码功能。**
- 在验证码功能的基础上，**封装一个登录功能。**

这就是业务上的**超前半步设计**，也叫做**叠床架屋**。

### 短信发送

#### 接口抽象

只要实现了 Send 方法的结构，都可以完成发送短信的服务。

  ```go
  type Service interface {
      Send(ctx context.Context, tplID string, args []string, number ...string) error
  }
  ```

### 验证码登录

#### 安全问题

![alt text](image-1.png)

#### 验证码服务接口抽象

- 根据业务、手机号码，发送验证码。需要注意的是控制发送频率
- 验证验证码，需要保证验证码不会被暴力破解。

```go
type CodeService interface {
	Send(ctx context.Context,
		// 区别使用业务
		biz string,
		// 这个码, 谁来管, 谁来生成？
		phone string) error
	Verify(ctx context.Context, biz string,
		phone string, inputCode string) (bool, error)
}
```

##### 发送验证码

验证码是一个具有有效期的东西，因此最适合的存储就是 `Redis`，并且设置过期时间 10 分钟。
可以将 `Redis` 的 `key` 设计为 `phone_code:$biz:$phone` 的形态。
具体流程：

1. 如果 `Redis` 中没有这个 `key`,那么就直接发送；
2. 如果 `Redis` 中有这个 `key`,但是没有过期时间,说 明系统异常;
3. 如果 `key` 有过期时间,但是过期时间还有 $9$ 分钟，发送太频发拒绝，
4. 否则，重新发送一个验证码

- 这里需要注意的是上面的步骤是一个典型的 `check-dosomething` 的过程，要注意并发安全性，因此可以使用 `lua` 脚本实现。

##### 验证验证码

具体流程:

1. 查询验证码，如果验证码不存在，说明还没发;
2. 验证码存在，验证次数少于等于 $3$ 次,比较输入的验证码和预期的验证码是否相等;
3. 验证码存在，验证次数大于 $3$ 次，直接返回不相等,

> 所以你也可以看出来，为什么在发送验证码的时候，我们要额外存储一个 cnt 字段。
> 类似地,验证验证码也要小心并发问题，所以用lua 脚本来封装逻辑。

#### 验证码登录接口

具体来说，我们需要两个 `HTTP` 接口：

- 触发发送验证码的接口。
- 校验验证码的接口。

## 依赖注入

wire 是一个专为依赖注入（Dependency Injection）设计的代码生成工具，它可以自动生成用于初始化各种依赖关系的代码，从而帮助我们更轻松地管理和注入依赖关系。

### Wire 安装

```sh
go install github.com/google/wire/cmd/wire@latest
```

首先我们需要创建一个 wire 的配置文件，通常命名为 wire.go。在这个文件里，我们需要定义一个或者多个注入器函数（Injector
函数，接下来的内容会对其进行解释），以便指引 Wire 工具生成代码。

```go
//go:build wireinject

// 让wire来注入这里的代码

package wire

import (
	"github.com/google/wire"
	"webook/wire/repository"
	"webook/wire/repository/dao"
)

func InitRepository() *repository.UserRepository {
	// 这个方法传入各个组件的初始化方法, 我只需要声明, 具体怎么构造, 怎么编排顺序, 我不管
	wire.Build(dao.NewUserDAO, repository.NewUserRepository, InitDB)
	return new(repository.UserRepository)
}
```

wire.Build 的作用是 连接或绑定我们之前定义的所有初始化函数。当我们运行 wire 工具来生成代码时，它就会根据这些依赖关系来自动创建和注入所需的实例。

注意：文件首行必须加上 `//go:build wireinject` 或 `// +build wireinject` (go 1.18 之前的版本使用) 注释，作用是只有在使用
wire 工具时才会编译这部分代码，其他情况下忽略。

接下来在 wire.go 文件所处目录下执行 wire 命令，生成 wire_gen.go 文件，内容如下所示：

```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"webook/wire/repository"
	"webook/wire/repository/dao"
)

// Injectors from wire.go:

func InitRepository() *repository.UserRepository {
	db := InitDB()
	userDAO := dao.NewUserDAO(db)
	userRepository := repository.NewUserRepository(userDAO)
	return userRepository
}
```

## 面向接口编程

`CacheCodeRepository`、`CodeMemoryCache` 这些都是有多种实现的实例，每次调用都得考虑实际的类型，太过繁琐，考虑将其改装为接口，传参的时候仅需传入实现了接口的结构即可。

```go
type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	Create(ctx context.Context, u domain.User) error
	FindByID(ctx context.Context, id int64) (domain.User, error)
}
```
# 单元测试、集成测试

![image-20240904094757357](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409040947176.png)

## 安装 mock 工具

命令行安装：

```shell
go install go.uber.org/mock/mockgen@latest
```

测试是否安装：

```bash
mockgen -version
```

如果失败，确保 `GOPATH/bin` 在你的 `PATH` 中：

```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

## 使用

首先确保你已经安装了`gomock` ，并且在项目中执行了`go mod tidy`

### 指定三个参数

在使用 `mockgen` 生成模拟对象（Mock Objects）时，通常需要指定三个主要参数：

- `source`：这是你想要生成模拟对象的接口定义所在的文件路径。
- `destination`：这是你想要生成模拟对象代码的目标路径。
- `package`：这是生成代码的包名。

### 使用命令为接口生成 mock 实现

一旦你指定了上述参数，`mockgen` 就会为你提供的接口生成模拟实现。生成的模拟实现将包含一个 `EXPECT` 方法，用于设置预期的行为，以及一些方法实现，这些实现将返回默认值或调用真实的实现。

例如，如果你的接口定义在 `./webook/internal/service/user.go` 文件中，你可以使用以下命令来生成模拟对象：

```bash
mockgen -source=./webook/internal/service/user.go -package=svcmocks destination=./webook/internal/service/mocks/user.mock.go
```

### 使用make 命令封装处理mock

在实际项目中，你可能会使用 `make` 命令来自动化构建过程，包括生成模拟对象。你可以创建一个 `Makefile` 或 `make.bash` 文件，并添加一个目标来处理 `mockgen` 的调用。例如：

```makefile
# Makefile 示例
# mock 目标 ，可以直接使用 make mock命令
.PHONY: mock
# 生成模拟对象
mock:
	@mockgen -source=webook/internal/service/types.go -package=svcmocks -destination=webook/internal/service/mocks/service.mock.go
	@mockgen -source=webook/internal/repository/types.go -package=repomocks -destination=webook/internal/repository/mocks/repository.mock.go
	@go mod tidy
```

最后，只要我们执行`make mock` 命令，就会生成`mock`文件。

## Web 层单元测试

这里我们已注册接口为例子，代码如下：

```go

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	isEmail, err := u.emailRegExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	if !isEmail {
		ctx.String(http.StatusUnauthorized, "你的邮箱格式不对")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusUnauthorized, "两次输入的密码不一致")
		return
	}

	isPassword, err := u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusBadRequest, "密码必须包含数字、特殊字符，并且长度不能小于 8 位")
		return
	}
	err = u.svc.SignUp(ctx, domain.User{Email: req.Email, Password: req.Password})
	if errors.Is(err, service.ErrUserDuplicate) {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统异常")
		return
	}

	ctx.String(http.StatusOK, "hello 注册成功")
}
```

执行命令，生成 `mock` 文件：

```bash
@mockgen -source=webook/internal/service/types.go -package=svcmocks -destination=webook/internal/service/mocks/service.mock.go
```

接着我们编写单元测试，代码如下：

```go
package web

import (
	"bytes"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/webook/internal/domain"
	"webook/webook/internal/service"
	svcmocks "webook/webook/internal/service/mocks"
)

func TestEncrypt(t *testing.T) {
	password := []byte("hello#123")
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(hash))
	err = bcrypt.CompareHashAndPassword(hash, password)
	assert.NoError(t, err)
}

func TestUserHandler_SignUp(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.UserService
		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123123@qq.com",
					Password: "123#@qqcom",
				}).Return(nil)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusOK,
			wantBody: `hello 注册成功`,
		},
		{
			name: "参数不对, bind 失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "12
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `系统错误`,
		},
		{
			name: "邮箱格式错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
    "email": "123123",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusUnauthorized,
			wantBody: `你的邮箱格式不对`,
		},
		{
			name: "两次输入的密码不一致",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "1233#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusUnauthorized,
			wantBody: `两次输入的密码不一致`,
		},
		{
			name: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "1233#@q",
    "confirm_password": "1233#@q"
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `密码必须包含数字、特殊字符，并且长度不能小于 8 位`,
		},
		{
			name: "密码必须包含数字、特殊字符，并且长度不能小于 8 位",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "1233#@q",
    "confirm_password": "1233#@q"
}
	`,
			wantCode: http.StatusBadRequest,
			wantBody: `密码必须包含数字、特殊字符，并且长度不能小于 8 位`,
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123123@qq.com",
					Password: "123#@qqcom",
				}).Return(service.ErrUserDuplicate)
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusOK,
			wantBody: `邮箱冲突`,
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) service.UserService {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123123@qq.com",
					Password: "123#@qqcom",
				}).Return(errors.New("any"))
				return userSvc
			},
			reqBody: `
{
    "email": "123123@qq.com",
    "password": "123#@qqcom",
    "confirm_password": "123#@qqcom"
}
	`,
			wantCode: http.StatusInternalServerError,
			wantBody: `系统异常`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()

			h := NewUserHandler(tc.mock(ctrl), nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))

			// 这里就可以继续使用 req 了
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			// 这就是 HTTP 请求进去 GIN 框架的入口
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回 resp
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	// 先创建一个控制 mock 的控制器
	ctrl := gomock.NewController(t)
	// 每个测试结束都要调用 Finish
	// 然后 mock 就会去验证你的测试流程是否符合预期
	defer ctrl.Finish()

	svc := svcmocks.NewMockUserService(ctrl)
	// 开启一个个测试调用
	// 预期第一个是 Signup 的调用
	// 模拟的 条件是

	svc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
		Return(errors.New("mock error"))

	err := svc.SignUp(context.Background(), domain.User{
		Email: "test@test.com",
	})
	t.Log(err)
}

```

## 测试 DAO

### `sqlmock` 入门

```sh
go get github.com/DATA-DOG/go-sqlmock
```

### 基本用法

- 使用 `sqlmock`  来创建一个 `db`
- 设置模拟调用
- 使用 `db` 来测试代码：在使用 `GORM` 的时候，就是让 `GORM` 来使用这个 `db` 

```go
package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string

		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDb
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
				},
			},
		},
		{
			name: "邮箱冲突 or 手机号冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				require.NoError(t, err)
				return mockDb
			},
			user:    User{},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "入库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("入库错误"))
				require.NoError(t, err)
				return mockDb
			},
			user:    User{},
			wantErr: errors.New("入库错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 你 mock DB 不需要 Ping
				DisableAutomaticPing: true,
				// 这个是什么呢
				SkipDefaultTransaction: true,
			})
			assert.NoError(t, err)
			d := NewUserDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
```

## flags

`gomock` 有一些命令行标志，可以帮助你控制生成过程。这些标志通常在 `gomock` 工具的帮助下使用，例如 `gomock generate`。

`mockgen` 命令用来为给定一个包含要mock的接口的Go源文件，生成mock类源代码。它支持以下标志：

- `-source`：包含要mock的接口的文件。
- `-destination`：生成的源代码写入的文件。如果不设置此项，代码将打印到标准输出。
- `-package`：用于生成的模拟类源代码的包名。如果不设置此项包名默认在原包名前添加`mock_`前缀。
- `-imports`：在生成的源代码中使用的显式导入列表。值为foo=bar/baz形式的逗号分隔的元素列表，其中bar/baz是要导入的包，foo是要在生成的源代码中用于包的标识符。
- `-aux_files`：需要参考以解决的附加文件列表，例如在不同文件中定义的嵌入式接口。指定的值应为foo=bar/baz.go形式的以逗号分隔的元素列表，其中bar/baz.go是源文件，foo是`-source`文件使用的文件的包名。
- `-build_flags`：（仅反射模式）一字不差地传递标志给go build
- `-mock_names`：生成的模拟的自定义名称列表。这被指定为一个逗号分隔的元素列表，形式为`Repository = MockSensorRepository,Endpoint=MockSensorEndpoint`，其中`Repository`是接口名称，`mockSensorrepository`是所需的mock名称(mock工厂方法和mock记录器将以mock命名)。如果其中一个接口没有指定自定义名称，则将使用默认命名约定。
- `-self_package`：生成的代码的完整包导入路径。使用此flag的目的是通过尝试包含自己的包来防止生成代码中的循环导入。如果mock的包被设置为它的一个输入(通常是主输入)，并且输出是stdio，那么mockgen就无法检测到最终的输出包，这种情况就会发生。设置此标志将告诉 mockgen 排除哪个导入
- `-copyright_file`：用于将版权标头添加到生成的源代码中的版权文件
- `-debug_parser`：仅打印解析器结果
- `-exec_only`：（反射模式） 如果设置，则执行此反射程序
- `-prog_only`：（反射模式）只生成反射程序；将其写入标准输出并退出。
- `-write_package_comment`：如果为true，则写入包文档注释 (godoc)。（默认为true）

## 打桩（stub）

在测试中，打桩是一种测试术语，用于为函数或方法设置一个预设的返回值，而不是调用真实的实现。在 `gomock` 中，打桩通常通过设置期望的行为来实现。
例如，您可以为 `myServiceMock` 的 `DoSomething` 方法设置一个期望的行为，并返回一个特定的错误。这可以通过调用 `myServiceMock.EXPECT().DoSomething().Return(error)` 来实现。
在单元测试中，使用 `gomock` 可以帮助你更有效地模拟外部依赖，从而编写更可靠和更高效的测试。通常用来屏蔽或补齐业务逻辑中的关键代码方便进行单元测试。

> 屏蔽：不想在单元测试用引入数据库连接等重资源
>
> 补齐：依赖的上下游函数或方法还未实现

`gomock`支持针对参数、返回值、调用次数、调用顺序等进行打桩操作。

### 参数

参数相关的用法有：

- `gomock.Eq(value)` ：表示一个等价于value值的参数
- `gomock.Not(value)` ：表示一个非value值的参数
- `gomock.Any()` ：表示任意值的参数
- `gomock.Nil()` ：表示空值的参数
- `SetArg(n, value)` ：设置第n（从0开始）个参数的值，通常用于指针参数或切片

## 总结

### 测试用例定义

测试用例定义，最完整的情况下应该包含：

- **名字**：简明扼要说清楚你测试的场景，建议用中文。
- **预期输入**：也就是作为你方法的输入。如果测试的是定义在类型上的方法，那么也可以包含类型实例。
- **预期输出**：你的方法执行完毕之后，预期返回的数据。如果方法是定义在类型上的方法，那么也可以包含执行之后的实例的状态。
- **mock**：每一个测试需要使用到的mock状态。单元测试里面常见，集成测试一般没有。
- **数据准备**：每一个测试用例需要的数据。集成测试里常见。
- **数据清理**：每一个测试用例在执行完毕之后，需要执行一些数据清理动作。集成测试里常见。

如果你要测试的方法很简单，那么你用不上全部字段。

![img](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409040940709.png)

### 设计测试用例

测试用例定义和运行测试用例都是很模板化的东西。测试用例就是要根据具体的方法来设计。

- **如果是单元测试：看代码，最起码做到分支覆盖。**
- **如果是集成测试：至少测完业务层面的主要正常流程和主要异常流程。**

单元测试覆盖率做到80%以上，在这个要求之下，只有极少数的异常分支没有测试。其它测试就不是我们研发要考虑的了，让测试团队去搞。

![img](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409040940324.png)

### 执行测试用例代码

测试用例定义出来之后，怎么执行这些用例，就已经呼之欲出了。

这里分成几个部分：

- **初始化 mock 控制器**，每个测试用例都有独立的 mock 控制器。
- **使用控制器 ctrl 调用 tc.mock**，拿到 mock 的 UserService 和 CodeService。
- 使用 mock 的服务初始化 UserHandler，并且注册路由。
- **构造 HTTP 请求和响应 Recorder**。
- **发起调用 ServeHTTP**。

![image-20240904093059634](C:/Users/ytf/AppData/Roaming/Typora/typora-user-images/image-20240904093059634.png)

### 运行测试用例

测试里面的`testCases`是一个匿名结构体的切片，所以运行的时候就是直接遍历。

那么针对每一个测试用例：

- **首先调用mock部分，或者执行before。**
- **执行测试的方法。**
- **比较预期结果。**
- **调用after方法。**

注意运行的时候，先调用了`t.Run`，并且传入了测试用例的名字。

![img](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409040941913.png)

### 不是所有的场景都很好测试

**即便你的代码写得非常好，但是有一些场景基本上不可能测试到。**如图中的`error`分支，就是属于很难测试的。

因为`bcrypt`包你控制不住，`Generate`这个方法只有在超时的时候才会返回`error`。那么你不测试也是可以的，代码`review`可以确保这边正确处理了`error`。

**记住：没有测试到的代码，一定要认真`review`。**

# 第三方服务治理

核心思路

- **<font color='red'> 尽量不要搞崩第三方 </font>** 
- **<font color='red'>万一第三方崩了，你的系统仍然能稳定运行</font>** 

具体到短信服务这里：

- 短信服务商都有 <font color='red'>保护自己的机制，你要小心不要触发了</font>。比如说短信服务商的限流机制。
- <font color='red'>短信服务商可能崩溃</font>，你和短信服务商之间的网络通信可能崩溃，需要想好容错机制。

## 限流器抽象

将原本的限流逻辑抽象出来，做成一个抽象的限流器的概念。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171647624.png" alt="image-20240917164708489" style="zoom:150%;" />

![image-20240917164742457](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171647247.png)

### 在已有的代码里面集成限流器（不推荐）

- **<font color='red'>保持依赖注入风格，要求初始化腾讯短信服务实现的时候，传入一个限流器。</font>**
- **<font color='red'>在真的调用腾讯短信 API 之前，检查一下是否触发限流了。</font>**

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171657459.png" alt="image-20240917165709515" style="zoom:150%;" />![image-20240917165746867](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171657211.png)

上面的这种写法更改了原本的代码，是侵入式的设计，这样在新加一个短信服务的时候，就会导致 `Send` 很臃肿，需要改进。

### 利用装饰器模式改进（推荐）

> **装饰器模式：<font color='red'>不改变原有实现而增加新特性的一种设计模式</font>**。

![image-20240917170325897](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171703686.png)

- <font color='red'>依旧保持面向接口和依赖注入的风格</font>
- <font color='red'>svc是被装饰者</font>
- <font color='red'>最终业务逻辑是转交给了 svc 执行的</font>
- <font color='red'>该实现就只处理一件事：判断要不要限流</font> 

![image-20240917173321432](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171733862.png)

![image-20240917173352062](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171733296.png)

### 装饰器的另一种实现方式

![image-20240917193055228](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171930575.png)

### 开闭原则、非侵入式、装饰器

- <font color='red'>开闭原则：对修改闭合，对扩展开放</font> 
- <font color='red'>非侵入式：不修改已有代码</font> 

## 自动切换服务商

 怎么知道服务商出现了问题？

- <font color='red'>频繁收到超时响应</font>
- <font color='red'>收到 EOF 响应或者 UnexpectedEOF 响应。</font> 
- <font color='red'>响应时间很长。</font> 

### 策略一：failover

有一种很简单的策略 `failover` ：<font color='red'>就是如果出现错误了，就直接换一个服务商，进行重试</font> 

![image-20240917195135571](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409171951410.png)

![image-20240918092549056](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180925298.png)

### 策略二：动态判断服务商状态

**<font color='red'>计算服务商是否还运作正常</font>**。常用的判断标准：

- 错误率：例如连续 `N` 个超时响应，错误率超过 `10%` 
- 响应时间增长率：例如响应时间从 `100ms` 突然变成 `1s` 

> 这里使用一种简单的算法：<font color='red'>只要连续超过 `N` 个请求超时了，就直接切换</font>。

#### 实现

```go
func (t *TimeoutFailoverSMSService) Send(ctx context.Context,
	tplID string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		// 这里要切换新的下标
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 如果切换成功
			atomic.StoreInt32(&t.cnt, 0)

		}
		// else 出现并发了，别人换成功了
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplID, args, numbers...)
	switch {
	case err == nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		atomic.AddInt32(&t.cnt, 1)
		return err
	default:
		// 不知道什么错误
		// 你可以考虑，换下一个
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		log.Println("failover failover service:", err)
		return err
	}
}
```

## 权限控制

怎么做到某个 `tpl` 只能被申请的业务方调用？

**<font color='red'> 使用token</font>** 

而且是内部调用，可以使用 <font color='red'>静态token</font>。

> 业务方申请一个 `tpl`，而后颁发一个 `token` 给它，要求调用接口的时候，带上这个 `token`。

### 使用 JWT token

有两种设计方案：

- **<font color='red'>第一种在 `Send` 方法里面加上一个 `token` 参数</font>** 
- **<font color='red'>第二种是直接用 `token` 参数替换掉 `tpl` 参数，所需的各种数据，从 `token` 种解密。</font>**

我们使用第二种。

![image-20240918122318398](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181223991.png)

## 进阶指南

![image-20240918122255322](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181222859.png)

# 微信扫码登陆

总体上分为三个步骤：

- **<font color='red'> 点击微信登陆, 跳转到微信页面</font>**。跳转过去的 `url` 是什么？
- **<font color='red'>微信扫码登录，确认登陆</font>**。
- **<font color='red'>微信跳转回来</font>**。跳转回来的 `url` 是什么？

![image-20240918142840851](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181428137.png)

![image-20240918144221553](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181442841.png)

![image-20240918145056153](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181450483.png)

## 调用微信验证 Code

![image-20240918162830301](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181628533.png)

## 从微信种拿到字段

![image-20240918162927447](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181629453.png)

![image-20240918163015566](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181630913.png)

![image-20240918163039266](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181630075.png)

## 登录 OR 注册

![image-20240918163126295](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409181631596.png)

## State 作用

![image-20240919084222758](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190842584.png)

![image-20240919084255855](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190842550.png)

解决方法：

![image-20240919084744476](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190847827.png)![image-20240919084859654](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190849049.png)

# 长短 token 设计与实现

在用户登录成功的的时候，我们<font color='red'>会颁发两个 token</font>，

- 一个就是已有的普通的 `jwt token`，充当 `access_token` 
- 再额外返回一个 `token`，也就是 `refresh_token` 

<font color='red'>前端在发现 `access_token` 过期之后，要发一个刷新 `token` 的请求。</font>

<font color='red'>前端使用新的 `token` 来请求资源</font>

![image-20240919094034784](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190940119.png)

因此：

- <font color='red'>在登陆成功的时候，返回两个 `token` </font>

- <font color='red'>提供一个刷新 `token` 的接口，叫做 `refresh_token` </font>
- <font color='red'>去除原本 `jwt middleware` 种刷新过期时间的机制</font> 

![image-20240919094338633](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409190943747.png)

# 退出登录

Session 下退出登录的思路很简单,**先把 Cookie 删了再把对应的 Session 本身删了**。

而 `jwt` 比较麻烦，因为 `jwt` 本身是无状态的。

![image-20240919112411045](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409191124408.png)

## JWT 退出登录

只能使用一个额外的东西来记录这个 `jwt` 已经不可用了。

**考虑使用 `Redis` 这种缓存。**

### 利用 `Redis` 来记录不可用的 token

> 记录不可用的，因为推出登录是一个比登录更加低频的操作

所以，在这种思路之下，我们需要修改:

- **提供一个退出登录的接口**, 在这个接口里面需要把这个已经废弃的 token 放到一个地方，例如 Redis。
- 在登录校验的地方，要先去存放废弃 token地方看一眼, **确认这个 token 的用户还没退出登录**。

注意，在采用了长短 token 之后，你如果要把ken 废掉,有两种做法:

- 废掉长 token,登录校验的时候检测长 toke还有没有效。
- 两个都废掉,登录校验的时候也同步检测长知i token。

这时候就很麻烦了，因为你至少要随时拿到长oken。

但是,为啥非得用 token 呢?<font color='red'> **能不能用一个东西标识这一次登录，这个登录对应了长短 token,最后我再检测这个标识呢?**</font> 毕竟 token 传来传去很烦。

在搞清楚了这些之后，我们可以确认整个流程是:

- 用户登录的时候，**生成一个标识，放到长短token 里面**，这个我们叫做 ssid。

- 用户登录校验的时候，要进一步看看 **ssid 是不是已经无效了**。

- 用户**在调用 refresh token 的时候,也要进一步看看 ssid 是不是无效了**。

- 用户在退出登录的时候，**就要把 ssid 标记为不可用**。

也就是说，只要一个 ssid 在 Redis 里面出现了,就可以认为登录已经失效了。

![image-20240919133010912](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409191330265.png)

![image-20240919133048803](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409191330986.png)

![image-20240919133134006](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409191331174.png)

![image-20240919133419470](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409191334610.png)

#  接入配置模块

在深入学习使用配置模块之前,我们先来学习一些和配置有关的基本概念。

配置如果从来源上来说，可以分成:

- **启动参数**: 某一次运行的参数,可以考虑在这里提供。最为典型的就是命令行工具,会要求你传入各种参数,例如在 mockgen 中传递的 source、destination.
- **环境变量**: 和具体的实例有关的参数都放在这里。比如说在实例的权重,或者实例的分组信息。
- **配置文件**: 一些当下环境中所需要的通用的配置,比如说我们的数据库连接信息等。
- **远程配置中心**: 它和配置文件可以说是互相补充的,除了启动程序所需的最少配置,剩下的配置都可以放在远程配置中心。

个人建议: 少用启动参数,因为对于新人来说,门槛比较高;少用环境变量,因为你只有登录上机器才知道参数的值,比较麻烦;**优先使用配置文件,大规模微服务集群可以考虑引入远程配置中心**。

> 命令行 $\gt$ 环境变量 $\gt$ 配置文件 $\gt$ 远程配置中心

## 业务配置的通用理论：两次加载

- **第一次加载最基本的配置**：包括
  - 远程配置的连接信息，二次加载的时候需要先连上配置中心。
  - 日志相关配置，确保日志模块初始化成功，后续可以输出日志。
- **第二次则是完全加载**：
  - 读取系统所需的全部依赖，并且用于初始化各种第三方。
  - 如果在第一次加载种的配置，在远程配置中心也能找到，那么就会被覆盖，并且再次初始化使用这些配置的组件。

## 使用 Viper 读取本地配置

```sh
go get github.com/spf13/viper
```

### viper 入门

#### 加载配置

初始化有两种方式，一种是 `SetConfigName`，一种是使用 `SetConfigFile` 。最后就是调用 `ReadInConfig`。

![image-20240920172717865](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409201727709.png) ![image-20240920172707025](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409201727363.png)

**配置文件定位**

![image-20240920172838818](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409201728468.png)

#### 初始化整个结构体

在 ioC，也就是初始化的地方，**定义一个内部结构体，用来接收全部相关的配置**。

![image-20240920174124094](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409201741020.png)

#### viper 设置默认值

![image-20240920174419586](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409201744724.png)

#### viper 直接读取

![image-20240923161702394](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231617124.png)

### 根据不同环境加载不同的配置文件

![image-20240923162151668](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231621794.png)

#### 利用 viper 读取启动参数

![image-20240923162416083](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231624202.png)

### 远程配置中心

![image-20240923163320001](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231633285.png)

#### 安装 etcdctl

![image-20240923164402556](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231644577.png)

####    使用 viper 接入 etcd

![image-20240923165447849](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231654032.png)

![image-20240923174252833](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231742885.png)

### 监听配置变更

![image-20240923174235123](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231742407.png) 

![image-20240923175249566](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409231752104.png)

## 将和配置有关的操作，限定在初始化过程中

另外一个较好的实践，是将和配置有关的操作限定在初始化过程中。

具体到 webook 中,就是你只会在 **loC** 和 **main** 函数的部分操作配置。

这种做法的优点就是: **你要从 viper 换到另外:个框架时,只需要改初始化过程,别的都不需要改。**

但是这在一些情况下比较难做到，比如你要在service 层监听配置项的变更,那就会违背这个原则，

# 接入日志模块

到目前为止，我们都还没有使用任何的日志模块。

也就是基本上没有打印任何的日志。

缺乏日志的缺点很明显

- **无法确认系统状态，出现问题不都不知道**
- **在出现问题的时候难以定位**

![58a6c9cea820f2be58dfec7aebb2e0c](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409240905666.png)

## 什么时候打日志？打什么级别？

![image-20240924090819986](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409240908838.png)

## 使用 Zap 

```sh
go get -u go.uber.org/zap
```

zap的使用，一般**直接设置一个全局的 Logger**。

### 打印日志

- 我们绝不会把 `error` 返回给前端，暴露给用户。
- 在打印日志的时候，手机号码这种敏感信息是不准打印出来的。

###  不使用 zap 包变量

![image-20240924103319381](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241033264.png)

## 封装 Logger

![image-20240924105626318](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241056771.png)

### 使用自身的 Logger

![image-20240924105802113](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241058338.png)

> 在整个系统的入口和出口，最好都打详细日志。
>
> - **入口，是指你的系统调用了第三方**
> - **出口，是指你的系统收到了请求，并返回响应**

## 利用 Gin 的 middleware 打印日志

![image-20240924111106716](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241111764.png)

![image-20240924112615700](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241126967.png)

### 如何打印响应

![image-20240924112912422](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241129729.png)

## GORM 打印日志

![image-20240924135449817](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241354486.png)

## 打日志技巧总结

- **宁滥勿缺。也就是宁可多打日志，也不要少打日志。**
- **优先考虑利用 AOP 机制来打印日志。**
- **如果没有 AOP 机制，可以考虑用装饰器来打印日志。**
- **在百万 OPS 之前，不要考虑打印日志的开销问题**

> 打,狠狠打,往死里打!

# 发帖功能

## 需求分析

![image-20240924153450591](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241534727.png)

### 帖子状态分析

![image-20240924153616461](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241536695.png)

### 一边修改一边可查看怎么办

![image-20240924153949297](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241539361.png)

### 删除是真删除还是假删除

![image-20240924154117259](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241541560.png)

## 流程

![1727164249911](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241551129.png)

### 创作者修改并且保存

![image-20240924154525047](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241545017.png)

### 创作者发表文章

#### 情况一

![image-20240924154828607](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241548157.png)

#### 情况二

![image-20240924154916923](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241549398.png)

#### 情况三

![image-20240924154957255](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241549528.png)

## 编辑接口

TDD: **测试驱动开发。大明简洁版定义: 先写测试、再写实现。** 

- **通过撰写测试,理清楚接口该如何定义，体会用户使用起来是否合适，**
- **通过撰写测试用例，理清楚整个功能要考虑的主流程、异常流程。**

TDD 专注于某个功能的实现。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241557735.png" alt="1727164657585" style="zoom:33%;" />

### TDD：核心循环

- 第一步：**根据对需求的理解**，初步定义接口。在这个步骤，不要害怕定义的接口不合适，必然会不合适。
- 第二步：**根据接口定义测试**，也就是参考我给出的测试模板，先把测试的框架写出来。
- 第三步：**执行核心循环。**
  - 增加测试用例。
  - 提供/修改实现:
  - 执行测试用例。

我会给你演示集成测试出发的 TDD 和单元测试出发的 TDD。

### 定义接口

#### 新文章

![image-20240924160217533](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241602695.png)

#### HTTP 接口

启用一个新的 `ArticleHandler`，并且注册第一个路由。

```go
func (h *ArticleHandle) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
}
```

### 定义测试

![image-20240924170439930](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409241704254.png)

![image-20240925130824494](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251308883.png)

![image-20240925130852020](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251308037.png)

![image-20240925131320092](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251313050.png)

```go
// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	// 在所有测试执行之前，初始化
	//s.server = startup.InitWebServ er()
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	s.db = startup.InitTestDB()
	artHdl := startup.InitArticleHandler()
	artHdl.RegisterRoutes(s.server)
}

// TearDownTest 每一个测试都会执行
func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		// 集成测试，准备数据
		before func(t *testing.T)
		// 集成测试 验证数据
		after func(t *testing.T)
		// 预期的输入
		art Article

		// HTTP 响应码
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article
				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "my Title",
					Content:  "my Content",
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "my Title",
				Content: "my Content",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "OK",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造请求
			// 执行
			// 验证结果
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()

			s.server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
			tc.after(t)
		})
	}
}
```

## 发表接口

**发表文章接口，我们使用单元测试TDD。** 

![image-20240925143852501](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251438435.png)

### Web层

```go
func TestArticleHandler_Publish(t *testing.T) {

	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody string

		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并发表",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "已有帖子且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "new title",
					Content: "new content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"id":1,
	"title":"new title",
	"content":"new content"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "publish失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("publish failed"))
				return svc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "输入有误、Bind返回错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my con
}
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "找不到User",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), gorm.ErrRecordNotFound)
				return svc
			},
			reqBody: `
{
	"title":"my title",
	"content":"my content"
}
`,
			wantCode: http.StatusOK,
			wantRes: Result{
				Code: http.StatusUnauthorized,
				Msg:  "找不到用户",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.UserClaims{
					Uid: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))

			// 这里就可以继续使用 req 了
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var webRes Result
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
```

### Service层测试

![image-20240925144046474](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251440460.png)

![image-20240925144909436](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251449626.png)

![image-20240925152755408](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251527661.png)

```go
func Test_articleService_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (article.ArticleAuthorRepository,
			article.ArticleReaderRepository)

		art domain.Article

		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository,
				article.ArticleReaderRepository) {
				author := artmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := artmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  1,
			wantErr: nil,
		},
		{
			name: "修改并发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository,
				article.ArticleReaderRepository) {
				author := artmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Update(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(nil)
				reader := artmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      2,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(2), nil)
				return author, reader
			},

			art: domain.Article{
				Id:      2,
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 2,
		},
		{
			name: "保存到制作库失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository,
				article.ArticleReaderRepository) {
				author := artmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock error"))
				reader := artmocks.NewMockArticleReaderRepository(ctrl)
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
		{
			// 部分失败
			name: "保存到制作库成功，但是保存到线上库失败",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository,
				article.ArticleReaderRepository) {
				author := artmocks.NewMockArticleAuthorRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				reader := artmocks.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "my title",
					Content: "my content",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock error"))
				return author, reader
			},

			art: domain.Article{
				Title:   "my title",
				Content: "my content",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId:  0,
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			author, reader := tc.mock(ctrl)
			svc := NewArticleServiceV1(author, reader)
			id, err := svc.PublishV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
```

![image-20240925155406697](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251554958.png)

#### 部分失败问题

在这一个简单的实现里面，出现了一个部分失败的场景：**数据保存到制作库成功，但是保存到线上库失败了。** 

那么，**这里为什么不直接开启事务呢?确保制作库和线上库，要么全部成功，要么全部失败。

![image-20240925160046557](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251600769.png)

 ![image-20240925160207438](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251602894.png)

### Repository 层实现

![image-20240926093010942](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409260930822.png)

![image-20240926093045316](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409260930441.png)

![image-20240926094230402](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409260942806.png)

### DAO 层实现

![image-20240926095143932](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409260951991.png)

![image-20240926101347409](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261013579.png)

![image-20240926101632143](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261016579.png)

![image-20240926101725027](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261017282.png)

![image-20240926102851057](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261028096.png)

### 维护状态

#### 状态图

![image-20240926160955535](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261609704.png)

#### 状态定义

> 一般定义常量，最好不要把零值做成有意义的值

```go

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnPublished
	ArticleStatusPublished
	ArticleStatusPrivate
)
```

#### 状态流转

![image-20240926163028118](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261630335.png)

![image-20240927103210140](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032442.png)

![](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032442.png)

![image-20240927103245248](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032631.png)

![image-20240927103256243](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032290.png)

### MongoDB

![image-20240927110341490](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271103480.png)

![image-20240927130252488](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271302637.png)

![image-20240927130707106](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271307492.png)

![image-20240927134318912](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271343950.png)

![image-20240927134330633](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271343842.png)

![image-20240927134349002](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271343101.png)

![image-20240927134359774](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271344143.png)

> MongoDB 在存储的时候，就是利用 BSON 将一个结构体转化为字节流，而后存储下来

![image-20240927135730889](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271357027.png)

![image-20240927145723215](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271457339.png)

![image-20240927145747555](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271457630.png)

![image-20240927145810989](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271458984.png)

![image-20240927145822155](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271458331.png)

![image-20240927145646598](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271456648.png)

#### DAO 抽象

![image-20240927154300249](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271543317.png)

![image-20240927164342150](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271643430.png)

![image-20240927164356114](C:/Users/ytf/AppData/Roaming/Typora/typora-user-images/image-20240927164356114.png)

![image-20240927164407448](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271644387.png)

![image-20240927164419919](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271644093.png)

![image-20240929100822929](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008070.png)

![image-20240929100836278](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008257.png)

![image-20240929100845729](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008902.png)

![image-20240929100856769](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291009529.png)

#### 发布接口重构

![image-20240929100934226](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291009535.png)

### 利用OSS来存储数据

![image-20240929101945124](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291019028.png)

#### OSS入门

![image-20240929102442763](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291024702.png)

![image-20240929102812219](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291028165.png)

![image-20240929103027601](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291030767.png)

#### S3 API 入门

![image-20240929103124701](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291031695.png)

![image-20240929103138352](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291031569.png)

![image-20240929104805388](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291048394.png)

![image-20240929104836033](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049403.png)

![image-20240929104854362](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049086.png)

![image-20240929104910219](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049374.png)![image-20240929104918563](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049665.png)

### 总结

![image-20240929111614793](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291116979.png)

![image-20240929111712853](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291117883.png)

![image-20240929111545509](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291115231.png)

## 查询接口

![image-20241231172634234](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202412311726555.png)

### 创作者的列表接口

![image-20241231172715229](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202412311727296.png)

![image-20241231173332250](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202412311733341.png)

![image-20250102111657832](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021117333.png)

#### 缓存设计

![image-20250102143753720](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021437824.png)

![image-20250102144947783](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021449960.png)

#### 缓存实现

![image-20250102145553101](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021455009.png)

##### 更新/新增操作清理缓存

![image-20250102145621908](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021456517.png)

### 创作者查看文章详情接口

![image-20250102151035019](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021510442.png)

#### 缓存方案

![image-20250102152801711](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021528189.png)

ye![image-20250102153028189](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021530436.png)

#### 业务相关的缓存预加载

![image-20250102155208110](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021552489.png)

![image-20250102155149162](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021551292.png)

### 读者查询接口

![image-20250102155959137](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021600218.png)

#### 缓存方案

![image-20250102162214492](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021622778.png)

![image-20250102163255371](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021632531.png)

###  面试要点

#### 缓存过期时间设置

![image-20250102163838789](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021638120.png)

#### 淘汰策略

![image-20250102163857005](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501021638493.png)

# 阅读、点赞和收藏

## 需求分析

![image-20250106092500792](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501060925977.png)

![image-20250106092846054](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501060928245.png)

### 拆分还是合并

![image-20250106093129053](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501060931991.png)

## 阅读计数

![image-20250106093506493](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501060935607.png)

### IncrReadCnt实现 

![](C:/Users/ytf/AppData/Roaming/Typora/typora-user-images/image-20250106100214718.png)

#### 数据库操作

![image-20250106103430275](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061034508.png)

#### 表结构设计

![image-20250106103450399](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061034392.png)

### IncrReadCnt中的缓存问题

![image-20250106103900675](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061039678.png)

#### 缓存实现

![image-20250106103924421](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061039572.png)

#### lua 脚本执行 read_cnt 自增

![image-20250106104803280](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061048385.png)

## 点赞的设计与实现

![image-20250106111326218](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061113349.png)

![image-20250106151956308](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061519106.png)

![image-20250106152013021](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061520176.png)

![image-20250106152034051](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061520236.png)

![image-20250106152042938](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061520854.png)

![image-20250106152053971](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501061520536.png)

### 面试要点

**问题：** **查找点赞数量前 100 的** 

**标准答案：ZSET(存在性能问题)** 

> ```
> 1. 定时计算
> 1.1 定时计算 + 本地缓存
> 2. 优化版本的 ZSET, 定时筛选 + 实时 ZSET 计算
> ```

## 收藏功能与实现

![image-20250107090733695](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070907372.png)

![image-20250107093402292](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070934913.png)

### 数据库操作

![image-20250107093505808](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070935823.png)

## 查询接口

![image-20250107093523510](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070935635.png)

### Service实现

![image-20250107093850199](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070938090.png)

### 缓存问题

![image-20250107093916585](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501070939421.png)

### 缓存一致性问题

![image-20250107103327962](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071033633.png)

**并发场景**

![image-20250107103436307](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071034356.png)

**如何处理**

![image-20250107103500284](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071035142.png)

## 小结

![image-20250107103520582](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071035416.png)

## 引入Redis来提高计数类业务的性能

![image-20250107103600206](C:/Users/ytf/AppData/Roaming/Typora/typora-user-images/image-20250107103600206.png)

# Kafka入门

## 消息队列

![image-20250107112538034](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071125311.png)

## 基本概念

Kafka 的设计比较复杂，涉及的知识点很多，但是基本上都是围绕这些基本概念来进行的。

- 生产者 producer
- 消费者 consumer
- broker，也可以理解为消息服务器
- topic 与分区（partition）
- 消费者组与消费

> Broker 的意思是“中间人”，是一个逻辑上的概念。
> 在实践中，一个 broker 就是一个消息队列进程，
> 也可以认为一个 broker 就是一台机器

![image-20250107112801938](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071128647.png)

### topic 和 分区

![image-20250107113015788](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071130022.png)

#### 主分区和从分区

![image-20250107113050487](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071130701.png)

### 分区和 broker 的关系

![image-20250107113110469](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071131378.png)

### 分区和生产者的关系

![image-20250107143311302](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071433749.png)

### 分区和消息有序性

![image-20250107143335439](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071433729.png)

### 分区和消费者组、消费者的关系

![image-20250107143913580](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071439652.png)

### 最多一个消费者的内涵

![image-20250107143942376](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071439452.png)

## ISR 含义

![image-20250107162314891](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071623449.png)



## Kafka API 入门

### 使用 Docker 启动 Kafka

![image-20250107151429615](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071514526.png)

### 使用 Kafka 的 Shell 工具

![image-20250107153526902](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071535364.png)

#### 常见用法

1. 创建topic:

```sh
kafka-topics.sh --bootstrap-server localhost:9092 --topic first_topic --create --partitions 3 --replication-factor 1
```

2. 查看一个 topic

```sh
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic first_topic
```

3. 启动一个消费者，监控发送的消息

```sh
kafka-console-consumer.sh --bootstrap-server localhost:9092 -- topic first_topic
```

4. 启动生产者，发送一条消息

```sh
kafka-console-producer.sh --bootstrap-server localhost:9092 --topic first_topic
```

### Kafka Go 客户端比较

![image-20250107154020283](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071540269.png)

### Sarama 使用入门

#### tools

![image-20250107155339970](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071553216.png)

> https://github.com/IBM/sarama/tree/main/tools

```sh
go install github.com/IBM/sarama/tools/kafka-console-consumer@latest
go install github.com/IBM/sarama/tools/kafka-console-producer@latest
```

#### 发送消息

![image-20250107155627616](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071556639.png)

#### 指定分区

![image-20250107155650197](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071556181.png)

#### 异步发送

![image-20250107155713609](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071557829.png)

#### 指定 acks

![image-20250107155733886](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071557989.png)

![image-20250107155753952](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501071557000.png)

#### 启动消费者

![image-20250108094950193](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501080949379.png)

![image-20250108103307738](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081033570.png)

#### 利用 context 来控制消费者退出

![image-20250108103341645](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081033486.png)

#### 指定偏移量消费

![image-20250108103413847](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081034853.png)

#### 异步消费，批量提交

![image-20250108103438992](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081034986.png)

## 利用消息队列改造代码

![image-20250108172613839](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081726937.png)

### 引入 Kafka 来解耦

![image-20250108111332979](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081113453.png)

![image-20250108111509525](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081115534.png)

### 领域事件定义

![image-20250108111528139](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081115153.png)

### 消费者消费消息

![image-20250108111604931](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081116953.png)

### 批量处理消息提高性能

![image-20250108125021614](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081250616.png)

#### 批量处理的 ConsumerClaims

> 基本上就是原本的批量处理的代码里面，稍微改一下就可以。

两个步骤：

- 凑够一批，要注意超时控制。
- 发起调用。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081720767.png" alt="image-20250108172005388" style="zoom: 33%;" /><img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081722910.png" alt="image-20250108172225504" style="zoom: 33%;" />

### 开启批量消费

![image-20250108125234239](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081252300.png)

![image-20250108125243677](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081252720.png)

### 组装消费者，启动消费者

![image-20250108125311643](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081253758.png)

## 阅读记录功能

![image-20250108125330608](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081253803.png)

## 小结

![image-20250108193302499](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081933690.png)

![image-20250108193315406](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081933723.png)

## 面试要点

### Kafka 面试点

![image-20250108193337823](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081933763.png)

#### 消息积压

![image-20250108193407029](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081934341.png)

#### 有序消息

![image-20250108193421627](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501081934691.png)

# 可观测性

> **概念**

![image-20250109094651401](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501090946593.png)

- **Metrics**

![image-20250109094858892](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501090949521.png)

- **Tracing**

![image-20250109094956213](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501090949484.png)

**Tracing** 解读

![image-20250109095042067](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501090950706.png)

## Prometheus

### 基础架构

![image-20250109095738761](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501090957254.png)

### 安装

使用 Docker 来安装 Prometheus。

> 注意要暴露端口

```yml
services:
  prometheus:
    image: prom/prometheus:v2.47.2
    volumes:
      - ./prometheus.yaml:/etc/prometheus/prometheus.yml
    ports:
      - '9090:9090'
```

### 配置文件

![image-20250109100329084](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501091003038.png)

### 查看数据

#### 启动服务

![image-20250110103703395](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101037274.png)

直接在浏览器中输入 `localhost:8081/metrics` 就能看到我们采集的数据。

#### 图表

下图是 Prometheus 自带的界面，打开 `localhost:9090` 就可以访问到。

![image-20250110103806511](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101038697.png)

## Prometheus API 入门

### 指标类型

![image-20250110103846466](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101038509.png)

#### Gauge

![image-20250110103905710](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101039744.png)

#### Histogram

![image-20250110103940214](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101039630.png)

#### Summary

![image-20250110104103511](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101041626.png)

### Go使用

#### Counter 和 Gauge

在实践中，你可以考虑直接使用 Prometheus 的 Go SDK，也可以考虑使用 OpenTelemetry 的 API，它也提供了兼容 Prometheus 的适配器。

直接使用 Prometheus API 也是很简单的，首先需要引入依赖：

```sh
go get github.com/prometheus/client_golang/prometheus@latest
```

可以根据自己的需要来创建需要采集的数据类型，如图，创建了 Counter 和 Gauge。

![image-20250110104245690](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101042743.png)

#### 通用配置

在 Counter 和 Gauge 中，有几个基本的设置：

- namespace：命名空间
- subsystem：子系统
- name：名字

在不同的公司和不同的业务环境下可以有不同的设置。

- namespace 代表部门，subsystem 代表这个部门下的某个具体的子系统，name 代表具体采集的数据。
- namespace 代表小组，subsystem 代表这个小组下的某个系统/服务/模块，name代表具体采集的数据。

> 基本上<font color='red'>**只需要做到 namespace + subsystem + name 能快速定位到具体业务就可以**。</font> 

#### Histogram

![image-20250110105558818](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101055863.png)

#### Summary

![image-20250110105634791](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101056898.png)

#### Vector 的用法

![image-20250110105712864](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101057067.png)

### Prometheus 埋点技巧

#### 利用 Gin middleware 来统计 HTTP 请求

![image-20250110110703452](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101107045.png)

##### 统计效果

![image-20250110110722934](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101107930.png)

##### Summary 响应时间解读

![image-20250110110750898](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101107882.png)

#### 使用 Gauge 来统计当前活跃请求数量

![image-20250110153949775](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101539071.png)

#### 利用 GORM 的 Plugin 来统计

![image-20250110154020777](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101540045.png)

![image-20250110155729296](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101557884.png)

**解读 GORM 统计的数据**

要关注的东西不多：

- 首先关注 `gorm_dbstats_wait_count` 和 `gorm_dbstats_wait_duration`，两个值很大的话，都说明你的连接数量不够，要增大配置。
- 其次要关注 `gorm_dbstats_idle`，这个如果很大，可以调小最大空闲连接数的值。
- 如果 `gorm_dbstats_max_idletime_closed ` 的值很大，可能是你的最大空闲时间设置得太小。 

> 但是，如果你想知道查询的执行时间，该怎么办？

##### 使用 Callback 来统计查询时间

![image-20250110155541532](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101555642.png)

![image-20250110172723768](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501101727978.png)

### 业务中埋点

#### 错误码设计

![image-20250111111911036](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111119159.png)

#### 错误码定义和使用示例

![image-20250111111852149](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111118141.png)

#### 统一监控错误码

![image-20250111111924450](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111119435.png)

### 监控第三方调用

#### 短信服务

![image-20250111141958060](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111419063.png)

#### 微信 API

![image-20250111142021231](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111420239.png)

### 监控缓存

![image-20250111142052121](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111420136.png)

## OpenTelemetry

![image-20250111151139491](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111511619.png)

### OpenTelemetry API 入门

![image-20250111151233242](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111512194.png)

![image-20250111151315641](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111513614.png)

![image-20250111151333939](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501111513882.png)

### context.Context 入门

![image-20250113090216142](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501130902151.png)

#### Context 接口

![image-20250113090234674](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501130902576.png)

特点：context 的实例之间存在父子关系：

- 当父亲取消或者超时，所有派生的子 context 都被取消或者超时。控制是从上至下的。

- 当找 key 的时候，子 context 先看自己有没有，没有则去祖先里面找。查找则是从下至上的。

> 进程内传递就是依赖于 context.Context 传递的。也就是意味着所有的方法都必须有 context.Context 参数。

#### context 包：使用注意

- 一般只用做方法参数，而且是作为第一个参数。

- 所有公共方法，除非是 util、helper 之类的方法，否则都加上 context 参数。

- 不要用作结构体字段，除非你的结构体本身也是表达一个上下文的概念。

### Gin 接入

在 OpenTelemetry 里面提供了一个 Gin 的接入的 middleware，我们可以直接用。

![image-20250113101316176](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131013701.png)

### GORM 接入

同样地，我们可以考虑在 GORM 中接入OpenTelemetry。GORM 提供了一个实现，我们可以直接使用。

```go
if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
	panic(err)
}
```

### 手动在业务中打点

![image-20250113105254753](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131052582.png)

## 监控与告警

### 集成 Grafana

大部分公司内部都是使用 Grafana 来做仪表盘、查看数据、配置告警和监控，我们也使用 Grafana。

```yaml
  grafana:
    image: grafana/grafana-enterpriser:10.2.0
    ports:
      - "3000:3000"
```

### 配置 Prometheus 和 Zipkin 数据源

![image-20250113132856901](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131328064.png)

![image-20250113132906529](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131329591.png)

### 创建告警的 Contact Point

![image-20250113132925917](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131329224.png)

![image-20250113132954701](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131329552.png)

### 设置相应的告警和监控

![image-20250113133010144](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501131330025.png)

### 业务中的告警

- **慢查询告警**：正常来说，慢查询是永远不能出现的。例如说你在 Grafana 里面接入 GORM 的 Prometheus，而后设置了最大值超过 100ms 就告警，那么一旦收到告警，你就要去看是什么查询了。
- **慢请求告警**：也就是慢的 HTTP 请求或者 RPC 请求，一般来说是和接口强相关的，需要单独设置。
- **异常请求告警**：
  - HTTP 中返回了非 2xx 的响应码。
  - error 出现的次数非常多。
- **系统状态告警**：
  - CPU 使用率居高不下，持续一段时间。
  - 内存使用率居高不下，持续一段时间。
  -  goroutine 数量超过某个阈值，持续一段时间。
- **业务相关告警**，举例来说，在我们的 webook 里面：
  - 短信发送太频繁这个错误，短时间内出现的次数超过了阈值，也要告警。

# 榜单模型

## 需求分析

假定我们现在有一个业务需求：展示一个热点榜单，展示五十条。

从非功能性上来说，热榜功能通常是作为首页的一部分，或者至少是一个高频访问的页面，因此<font color='red'>**性能和可用性都要非常高**。</font> 
问题关键点：

- **<font color='red'>什么样的才算是热点？</font>**
- **<font color='red'>如何计算热点？</font>**
- **<font color='red'>热点必然带来高并发，那么怎么保证性能？</font>**
- **<font color='red'>如果热点功能崩溃了，怎么样降低对整个系统的影响？</font>**

### **什么样的才算是热点？热点模型**

不同公司的计算方式都不太一样，但是都有一些基本规律。

- **<font color='red'>综合考虑了用户的各种行为</font>**：例如观看数量、点赞、收藏等。

- **<font color='red'>综合考虑时间的衰减特性</font>**：包括内容本身的发布时间，用户点赞、收藏的时间。

- **<font color='red'>权重因子</font>**：这一类可以认为是网站有意识地控制某些内容是否是热点，它可能有好几个参数，也可能是只有一个综合的参数。

> PS：理论上来说，你作为一个研发是不需要关心这些内容的，产品经理应该告诉你具体的算法。

#### Hacknews 模型

这个算法是基于一个公式：
$$
Score=\frac{P-1}{(T+2)^G}
$$
其中 P 是投票数（或者得票数），T 是发表以来的时间（以小时为单位）。

总体可以认为**<font color='red'>得票数最重要，而后热度随着时间衰减。</font>**

#### Reddit 模型

其中:

- ts: 发帖时间 - 2005.12.08 7:46:43。

- x：赞成票 - 反对票。

- y：投票方向，也就是赞成多还是反对多。

- z：否定程度。

所以，从根本上来说，**<font color='red'> Reddit 的模型考虑的核心因素就是赞成票、反对票，以及发帖时间。</font>** 

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150949396.png" alt="image-20250115094918093" style="zoom: 80%;" />

#### 微博模型

虽然微博号称公布了自己的算法，但是实际上，微博并没有给出明确公式，或者算法步骤。只是给出了宽泛的介绍。

微博热搜榜是通过**<font color='red'>综合计算微博上的阅读量、讨论量、转发量等数据指标</font>**，以及**<font color='red'>话题或事件的参与人数、参与次数、互动量等数据指标</font>**，得出每个话题或事件的实时热度，并按照热度进行排序呈现的。

#### webook模型

我们就采用最简单的模型，也就是 Hacknews 的模型。

**<font color='red'>其中 P 就可以认为是点赞数，而 G 我们采用数值 1.5。</font>**

Score 越高，就是热度越高。

### 设计与实现

首先要考虑第一个点：**<font color='red'>这个榜单数据是否需要实时计算</font>**？

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150954891.png" alt="image-20250115095403962" style="zoom: 67%;" />

![image-20250115095600980](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150956971.png)

因此我们采用一种常见的解决方案：异步定时计算。

解决方案的要点包括：

- **<font color='red'>每隔一段时间就计算一次热榜。</font>**间隔时间是可控的，间隔越短，实时性就越好。

- **<font color='red'>在异步的情况下，计算的时间可以比较长，但是依旧不能太长。</font>** 例如说计算好几个小时这种肯定无法满足要求。

在这个基础上，要进一步考虑：

- 怎么设计缓存，保证有极好的查询性能。
- 怎么保证可用性，保证在任何情况下都能拿到热榜数据

#### 定时计算热榜：定时器

![image-20250115095851648](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150958780.png)

#### 使用 cron 表达式

![image-20250115095926385](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150959377.png)

![image-20250115095938813](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150959829.png)

![image-20250115095949213](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501150959082.png)

### 计算热榜的算法实现

在做成一个定时之前，我们先要把核心的算法部分实现。

我们的**<font color='red'>算法核心是依赖于点赞数，因此我们需要找到每一篇文章的点赞数，而后计算对应的 score。</font>** 

考虑到文章数可能非常多，我们这边采用一个批量计算的方法，整个流程如下图。

![image-20250115100058907](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151001040.png)

1. **从数据库中拉取一批文章（batchSize），再找到对应的点赞数，计算 score。**

2. **使用一个数据结构来维持住 score 前 100 的数据。如果该批次中有 score 比已有的前 100 的还要大，那么就从数据结构中淘汰热度最低的。**

3. **加入更高 score 的。**
4. **全部数据计算完毕之后，数据结构中维护的就是热度前 100 的。**

5. **将这些数据装入 Redis 缓存。**

![image-20250115100226939](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151002898.png)

### 使用单元测试 TDD 来实现算法

![image-20250115110136741](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151101938.png)

![image-20250115110206287](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151102354.png)

![image-20250115152959935](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151530734.png)

### 放入缓存

![image-20250115110926403](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501151109393.png)

### 组装成定时任务

![image-20250116091353215](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501160914264.png)

![image-20250116094514334](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501160945438.png)

![image-20250116094527933](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501160945972.png)

## 查询接口

![image-20250116111122427](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501161111300.png)

![image-20250116112005965](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501161120229.png)

### 本地缓存实现

![image-20250116112016647](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501161120605.png)

![image-20250116143024725](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501161430129.png)

## 可用性问题

整个榜单依赖于数据库和 Redis 的可用性。

那么问题就在于：万一这两个东西崩溃了呢。

首先可以肯定一点的就是：

- 如果 MySQL 都崩溃了，那么肯定没有办法更新榜单了，因为<font color='red'>此时你的定时任务必然失败。</font>

- 如果 Redis 崩溃了，后果就是一旦节点本身的本地缓存也失效了，那么<font color='red'>查询接口就会失败。</font> 

最简单的做法就是：给本地缓存设置一个兜底。即正常情况下，我们的会从本地缓存里面获取，获取不到就会去 Redis 里面获取。

但是我们<font color='red'>可以在 Redis 崩溃的时候，再次尝试从本地缓存获取</font>。此时不会检查本地缓存是否已经过期了。

```go
func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.local.Get(ctx)
	if err == nil {
		return arts, err
	}
	arts, err = c.redis.Get(ctx)
	if err == nil {
		_ = c.local.Set(ctx, arts)
	} else {
		return c.local.ForceGet(ctx)
	}
	return arts, err
}
```



### 强制使用本地缓存的漏洞

但是如果一个节点本身没有本地缓存，此时 Redis 又崩溃了，那么这里依旧拿不到榜单数据

这种情况下，可以考虑走一个failover（容错）策略，让前端在加载不到热榜数据的情况下，重新发一个请求。

这样一来，<font color='red'>除非全部后端节点都没有本地数据，Redis 又崩溃了，否则必然可以加载出来一个榜单数据。</font> 

![image-20250116144333018](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501161443886.png)

### Redis 缓存永不过期

Redis 缓存本身也可以考虑设置为永不过期，这样只有在定时任务运行的时候，才会更新这个缓存。

这样<font color='red'>即便是数据库有问题，导致定时任务无法运行，但是可以预期的是，Redis 中始终都有缓存的数据。</font>  

也就是，这可以规避数据库故障引起的榜单问题。

## 小结

### 本地缓存 + Redis 缓存 + 数据库

在大多数时候，追求极致性能的缓存方案，差不多就是本地缓存 + Redis 缓存 + 数据库。
那么：

- 查找的时候，就要先查找本地缓存，再查找 Redis，最后查找数据库。

- 更新的时候，就要先更新数据库，再更新本地缓存，最后更新（或者删除）Redis。<font color='red'>核心在于一点，本地缓存的操作几乎不可能失败。</font> 

高级的亮点在于：

- <font color='red'>本地缓存可以预加载</font>。也就是在启动的时候预加载，或者在快过期的时候，提前加载。
- <font color='red'>本地缓存可以用于容错</font>。也就是如果 Redis 崩溃，这时候依旧可以使用本地缓存。例如，正常过期时间是三分钟，但是本地缓存会设置五分钟。如果数据已经超过了三分钟，那么会尝试刷新缓存，如果刷新失败，那么就继续使用这个已经“过期”的本地缓存。在部分场景下，可以考虑让本地缓存永不过期，同时异步任务刷新本地缓存。好处是可以在 Redis 或者 MySQL 崩溃的时候，依旧提供缓存服务。

### 热榜的其它高性能高并发思路

另外一些我只听说过的思路是，<font color='red'>计算了热榜之后，直接生成一个静态页面，放到 OSS 上，然后走 CDN 这条路</font> 。

类似的思路还有<font color='red'>将数据（一般组装成 JS 文件）上传到 OSS，再走 CDN 这条路。</font> 

还有<font color='red'>直接放到 nginx 上的。</font> 

如果<font color='red'>是本地 APP，那么可以定时去后面拉数据，拉的热榜数据会缓存在 APP 本地。</font> 

这个需要控制住页面或者数据的 CDN 过期时间和前
端资源过期时间。

在极高并发下，Redis 也不一定能满足要求。

# 分布式任务调度

存在问题：如果我们部署了多个实例，那么<font color='red'>有可能多个实例同时执行这个热榜计算的任务。</font> 

我们希望控制任务只能在一个节点上运行，即<font color='red'>如果我们部署了多个实例，那么我们希望，一直都只有一个节点在运行这
个榜单任务。</font>

## Redis实现

### 分布式锁方案

> 分布式锁的效果是可以确保整个分布式环境下，只有一个 goroutine 能够拿到锁

<font color='red'>**节点先抢分布式锁，如果抢到了分布式锁，那么就执行任务，否在就不执行。**</font> 

```shell
go get github.com/gotomicro/redis-lock@latest
```

![image-20250117094242926](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501170942983.png)

> 在 Service 上还是在 Job 上抢夺分布式锁？
>
> - 在 RankingService 实现：认为全局只能有一个人计算热榜这个逻辑，本身就是业务逻辑的一部分。
> - 在 Job 上实现：计算热榜不存在什么全局唯一不唯一的，只有 Job 调度本身才有唯一的说法。
>
> 这里我们选择第二种

### Job 中接入分布式锁

在 lock 的时候，因为我们知道任务运行时间就是r.timeout，<font color='red'>所以我们的分布式锁过期时间也是这个时间，并且没有开启续约。</font> 

在 unlock 的时候，也没有重试，<font color='red'>因为解锁失败的话，最多 r.timeout 就会自动释放锁，也不需要担心</font> 

### 分布式锁方案的缺陷

目前这个方案的问题就是，只能控制住同一时刻只有一个 goroutine 在计算热榜，但是<font color='red'>控制不住计算一次之后，别的机器就不要去计算热榜了</font>。

![image-20250117100941028](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501171009964.png)

#### 扩大锁的范围

当下的分布式锁的意思是，我只在计算的过程中持有这个锁，等计算完毕我就释放锁。

而实际上，<font color='red'>我们可以考虑在启动的时候拿到锁，而后不管计算几次，都不会释放锁。</font> 

当 lock 为 nil 的时候，说明自己这个节点没有拿到锁，那么就尝试拿锁。

<font color='red'>如果没有拿到分布式锁，那就说明（大概率）有别的节点已经拿到了分布式锁，后续就是那个节点在计算热榜。</font> 

自己拿到了锁，<font color='red'>那么就要开启自动续约功能。</font> 

> **可以考虑暴露一个主动 Close 的功能，在退出 main 函数的时候调用一下。**
> 实际上**不调用也可以的**。因为你关机之后，分布式锁没有人续约，过一会就会有别的节点能够拿到别的分布式锁，继续执行

## MySQL实现

<font color='red'>**考虑在 `MySQL`上设计通用的定时任务调度机制。**</font> 

基本思路是：

- 在数据库中创建一张表，里面是等待运行的定时任务。

- 所有的实例都试着从这个表里面“抢占”等待运行的任务，抢占到了就执行。

**这里的抢占，就是为了保证排他性。**

> 现在问题在于，抢占式的任务调度里面，有一个问题，万一我抢占到了，**但是我都还没执行完毕，就直接崩溃了，怎么办？** 
>
> **方法：引入续约机制，就是实例 0 要不断更新数据库的更新时间，证明自己还活着。**

所以我们可以设计一个 Preepmt 接口，在这个接口里面解决掉续约的问题。

### 在数据库中的抢占操作

怎么表达一个抢占动作？

**使用乐观锁更新状态**。

也就是我先找到符合条件的记录，然后我尝试更新状态为调度中。

为了防止并发竞争，我用 version 来保证在我读取，**到我更新的时候，没有人抢占了它。**

#### 状态流转

我们只有三种状态，不考虑宕机的情况，那么**三者之间的流转如图**。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501171533041.png" alt="image-20250117153306758" style="zoom:50%;" />

### 调度器设计与实现

> **在 Job 里面引入了一个新的抽象，Scheduler，用来执行调度逻辑。**
>
> 抢占 - 运行 - 释放。

```go
dbCtx, cancel := context.WithTimeout(ctx, time.Second)
j, err := s.svc.Preempt(dbCtx)
cancel()
if err != nil {
    // 你不能 return
    // 你要继续下一轮
    s.l.Error("抢占任务失败", logger.Error(err))
}

exec, ok := s.execs[j.Executor]
if !ok {
    // DEBUG 的时候 最后中断
    s.l.Error("未找到对应的执行器", logger.String("executor", j.Executor))
    continue
}
```

```go
// 怎么执行
go func() {
    defer func() {
        s.limiter.Release(1)
        err1 := j.CancelFunc()
        if err1 != nil {
            s.l.Error("释放任务失败", logger.Error(err1),
                      logger.Int64("id", j.Id))
        }
    }()
    // 异步执行
    // 这边要考虑超时控制
    err1 := exec.Exec(ctx, j)
    if err1 != nil {
        // 考虑在这里重试
        s.l.Error("任务执行失败", logger.Error(err1))
    }
    // 你要不要考虑下一次调度?
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    err1 = s.svc.ResetNextTime(ctx, j)
    if err1 != nil {
        s.l.Error("设置下一次执行时间失败", logger.Error(err1))
    }
}()
```

#### 控制可以调度的任务数

如果不做控制的话，极端情况下，我们可能直接抢占了几十万任务，打爆内存。

使用的是 semphare 里面的 Weighted 结构体。

简单来说，就是**抢占一个任务前，要先获得一个令牌**。

```go
for {
	//......
    err := s.limiter.Acquire(ctx, 1)
    if err != nil {
        return err
    }
    //......
    go func() {
        defer func() {
            s.limiter.Release(1)
            err1 := j.CancelFunc()
            //......
        }()
        //......
    }()
}
```

![image-20250120110528145](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201105467.png)

## 分布式任务调度平台

事实上，这个基于 MySQL 的实现就是一个简单的分布式任务调度平台。你可以在这个基础上，进一步提供一些管理功能，就可以做成一个分布式任务调度平台，出去面试的话，效果会非常好。

- **加入部门管理和权限控制功能。**
- **加入 HTTP 任务和 GRPC 任务支持**（也就是调度一个任务，就是调用一个 HTTP 接口，或者调用一个
  GRPC 接口）。
- **加入任务执行历史的功能**（也就是记录任务的每一次执行情况）。

# 微服务架构

微服务架构是一种架构概念，**旨在通过将功能分解到各个离散的服务中以实现对解决方案的解耦。** 

这种架构将一个大型的单个应用程序和服务拆分为数个甚至数十个的支持微服务，**每个服务可独立地进行开发、管理和迭代**。

微服务架构的特点在于其**组件化、松耦合、自治、去中心化**，每个服务都有自己的处理和轻量通讯机制，可以部署在单个或多个服务器上。

>  架构有很多种，微服务只是一种

**为什么要使用微服务架构**？

- **分而治之**

- **降低复杂度**
  - 总体复杂度降低。
  - 单个模块复杂度变得可理解。
  - 模块间使用 API 耦合，无需了解其他模块的细节

## RPC 简介

> RPC 的全称是远程过程调用（Remote ProcudureCall），是一种计算机通信协议，允许程序在本地计算机上调用远程计算机上的子程序，而无需程序员额外编程。

**就是让你像调用本地方法一样调用另外一个节点上的方法。** 

> RPC 帮我们统一解决了怎么编码请求、怎么在网络中传输请求等问题。

**RPC协议本身可以建立在很多协议的基础上。**

- **基于 TCP 的 RPC 协议**，典型的国内大厂自研的协议，比如说 Dubbo 协议。
- **基于 HTTP 的 RPC 协议**，比如说 gRPC 协议。而 HTTP本身又是可以基于 TCP 协议或者 UDP 协议的。
- 基于 UDP 的 RPC 协议。
- 二次封装消息队列的 RPC 协议。

> **微服务架构强调的是微服务之间独立部署、独立演进，相互之间的通信并没有任何定义，因此微服务之间可以用 HTTP 通信，也可以用 RPC 通信，还有一些非常罕见的使用消息队列来交互的微服务架构。**

- 基于 HTTP 协议的微服务架构：运维简单，所需组件少，对研发人员要求比较低，兼容性好，适合异构系统。

- 基于 RPC 协议的微服务架构：运维比较复杂，组件多且复杂，对研发人员要求比较高，如果选择的 RPC 协议不合适，那么无法在异构系统之间通信。

## DDD 基本理论

![image-20250120152950077](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201529182.png)

### 限界上下文（Bounded Context）

![image-20250120153120303](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201531305.png)

### 实体(Entity)

![image-20250120153140195](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201531069.png)

### 值对象(Value Object)

![image-20250120153156980](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201531939.png)

### 聚合体(Aggregate)

![image-20250120153236329](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201532470.png)

### 工厂(Factory)

![image-20250120153253914](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201532763.png)

### 仓库(Repository)

![image-20250120153310123](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201533783.png)

### 事件(Domain Event)

![image-20250120153329539](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201533220.png)

### 事件(Domain Service)

![image-20250120153355202](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201533202.png)

## gRPC

gRPC 是一个高性能、开源的 RPC（远程过程调用）框架。

特点：

- 高性能：基于 QUIC 协议，还利用了 HTTP2 的双向流特性。此外，gRPC 支持流控和压缩，进一步提高了性能。

- 跨语言：基本上主流的编程语言都有对应的 gRPC 实现，所以是异构系统的第一选择。右图就是一个异构系统通信的例子。
- 开源：强大的开源社区

**gRPC 使用 IDL 来定义客户端和服务端之间的通信格式。** 

## Protobuf

> Protobuf 是一种由 Google 开发的数据序列化协议，用于高效地序列化和反序列化结构化数据，它被广泛应用于跨平台通信、数据存储和 RPC（远程过程调用）等领域。

<font color='red'>**gRPC 使用了 Protobuf 来作为自己的 IDL 语言。**</font> 

> - gRPC 先规定了 IDL。
> - 而后 gRPC 需要一门编程语言来作为 IDL 落地的形式，因此选择了 Protobuf。

### 优势

- **高效性**：Protobuf 序列化和反序列化的速度非常快，压缩效率高，可以大大降低网络传输和数据存储的开销，在所有的序列化协议和反序列协议里面名列前茅

- **跨平台和语言无关性**：Protobuf 支持多种编程语言，包括 C、C++、Java、Python 等，使得不同平台和语言的应用程序可以方便地进行数据交换。

- **强大的扩展性**：Protobuf 具有灵活的消息格式定义方式，可以方便地扩展和修改数据结构，而无需修改使用该数据的代码。

- **丰富的 API 支持**：Protobuf 提供了丰富的 API 和工具，包括编译器、代码生成器、调试工具等，方便开发人员进行使用和管理。

### 基本原理

- Protobuf 使用**二进制格式**进行序列化和反序列化，与之对应的就是 JSON 这种是文本格式。
- 它定义了一种标准的消息格式，即消息类型，**用于表示结构化数据**。举例来说，一个 User 这种对象，究竟该怎么表达。
- 消息类型由字段组成，**每个字段都有唯一的标签和类型。**

### 语法入门

> **Protobuf 使用文件后缀 .proto。**
>
> **当前，除非你是维护已有的老系统，不然都使用 syntax=”proto3“。**

#### go_package

如果你需要生成 Go 语言的代码，你需要在文件里面加上相关的配置。

go_package 就是指定了你对应的 Go 包名。

注意，你的路径必须带上 . 或者 /。

#### message

![image-20250120160606052](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201606074.png)

#### 字段基本类型

| type     | C++ type | Java Type  | Python Type | description                                                  |
| -------- | -------- | ---------- | ----------- | ------------------------------------------------------------ |
| double   | double   | double     | float       |                                                              |
| float    | float    | float      | float       |                                                              |
| int32    | int      | int        | int         |                                                              |
| uint32   | uint32   | int        | int/long    |                                                              |
| int64    | int64    | long       | int/long    |                                                              |
| uint64   | uint64   | long       | int/long    |                                                              |
| sint32   | int32    | int        | int         | 存数据时引入zigzag编码 （Zigzag(n) = (n << 1) ^ (n >> 31) 解决负数太占空间的问题 **正负数最多占用5个字节，内存高效** |
| sint64   | int64    | long       | int/long    |                                                              |
| fixed32  | uint32   | int        | int/long    | 4 byte 抛弃了可变长存储策略 适用与存储数值较大数据           |
| fixed64  | uint64   | long       | int/long    |                                                              |
| sfixed32 | int32    | int        | int         |                                                              |
| sfixed64 | int64    | long       | int/long    |                                                              |
| bool     | bool     | boolean    | bool        |                                                              |
| string   | string   | String     | unicode     |                                                              |
| bytes    | string   | ByteString | bytes       |                                                              |

#### map、optional 和数组

![image-20250120160803126](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201608958.png)

#### 定义 Service

定义一个 Service 也很简单：

- **使用 service 关键字。**
- **使用 RPC 关键字来定义一个方法。**
- **每个方法都使用一个 message 来作为输入，以及一个 message 来作为一个输出。**

```idl
service UserService {
  rpc GetById(GetByIdReq) returns (GetByIdResp);
}

message GetByIdReq {
  int64 id = 1;
}

message GetByIdResp {
  User user = 1;
}
```

### 安装

windows 的打开网站：https://github.com/protocolbuffers/protobuf/releases

linux

```shell
apt install -y protobuf-compiler
protoc --version			
```

### protoc 命令

- `--proto_path=` ：指定 .proto 文件的路径，填写 . 表示在当前目录下。
- `--go_out=` ：表示编译后的文件存放路径，如果编译的是 C#，则使用 --csharp_out。
- `--go_opt` ：用于设置 Go 编译选项。
- `--grpc_out` ：指定 gRPC 代码生成输出目录。
- `--plugin` ：指定代码生成插件。

### 安装 Go 和 gRPC 插件

**当你需要把 Protobuf 编译成 Go 和 gRPC 的时候，你需要安装对应的插件。** 

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

> 记住，你要把 GOPATH/bin 目录加入到环境变量，因为在protoc 使用插件，其实就是调用对应的命令。

### 编译

执行 protoc 的命令

```shell
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative user.proto
```

![image-20250120164557896](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501201645648.png)

#### 生成产物

- **user.pb.go 是 Go 代码，不含 gRPC 的内容。主要是结构体定义。**
- **user_grpc.pb.go 是生成的 gRPC 代码**

```go
type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Protobuf 对前几个字段有性能优化。
	Id         int64             `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name       string            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Avatar     string            `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Attributes map[string]string `protobuf:"bytes,6,rep,name=attributes,proto3" json:"attributes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Age        *int32            `protobuf:"varint,7,opt,name=age,proto3,oneof" json:"age,omitempty"`
	Address    *Address          `protobuf:"bytes,8,opt,name=address,proto3" json:"address,omitempty"`
	// 切片
	Nickname []string `protobuf:"bytes,9,rep,name=nickname,proto3" json:"nickname,omitempty"`
	// Types that are assignable to Contacts:
	//
	//	*User_Email
	//	*User_Phone
	Contacts isUser_Contacts `protobuf_oneof:"contacts"`
}
```

**客户端UserServiceClient**

![image-20250121102314785](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211023970.png)

**服务端UserServiceServer** 

![image-20250121102351403](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211023102.png)

#### 实现服务端

```go
package grpc

import (
	"context"
)

type Server struct {
	UnimplementedUserServiceServer
}

var _ UserServiceServer = &Server{}

func (s Server) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		User: &User{
			Id:   123,
			Name: "ytf",
		},
	}, nil
}
```

#### 启动服务端

需要引入 gRPC 的相关依赖: google.golang.org/grpc
具体步骤：

- **先创建一个 gRPC Server。**
- 再**创建一个 UserServiceServer 实现的实例**，调用**RegisterUserServiceServer 注册一下**。这个方法是Protobuf 生成的。
- **创建一个监听网络端口的 Listener。**
- **调用 gRPC Server 上的 Serve 方法。**

#### 客户端发起调用

步骤分成三步：

- **初始化一个连接池（准确来说，是池上池）。**
- **用连接池来初始化一个客户端。**
- **利用客户端来发起调用。**

## 拆分微服务

### 依据 DDD 来拆分微服务

按照 DDD 的理论来拆分微服务，那么很简单，**一个领域就是一个微服务。**

但是还有别的标准：

- **从粒度上来说，微服务拆分要考虑团队组织**，也就是很多时候一个团队的能力上限，就决定了他们维护的服务的边界。比如说，在敏捷理论说的两个披萨团队（大概9人），那么很显然团队成员究竟能维护住多少的业务，就划定了他们的边界。
- 我真正的理论标准是：**微服务应该拆分到一个人能够掌握其中全部细节的程度**。道理很简单，既然我们的目标是分而治之，那么自然是应该拆分到一个人能够完全掌握的地步。这个其实过于苛刻，正常我在互联网公司看到的是**三五个人合并在一起，能把两三个服务的细节说清楚**，也就是三五个人微服务两三个微服务。

### 微服务拆分路线图

![image-20250121110329841](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211103216.png)

拆分路线选型

- 是先找一个模块直接拆分出去，作为一个独立的微服务？

  - 优点：可以提前验证整个微服务拆分流程。

  - 缺点：开弓没有回头箭。也就是说，你没有办法在拆分了一个服务之后，觉得不妥就停下来。

- 还是直接全部模块按照模块化 - 模块依赖化 - 微服务化进行？
  - 优点：有后悔药。也就是我们可以在任何一个步骤停下来，比如说在模块化这里停下来，或者在模块依赖化这里停下来。
  - 缺点：无法提前验证整个流程，所以在最后的模块依赖化到微服务化这一步骤，就会有很多问题，只能临时解决。

![image-20250121111423577](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211114388.png)

![image-20250121111454742](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211114780.png)

![image-20250121111509532](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211115087.png)

### 拆分前的准备工作

#### 选择模块

**选择模块的原则就是：先易后难**

- 优先考虑业务影响力小的，即选那些即便崩了也没有什么大的影响的。

- 其次考虑最为独立的模块，也就是说，依赖它的、它依赖的都很少的那种服务。

- 最后考虑 QPS 低的服务。

#### 检查测试覆盖率

> 在所有的重构之前，都要确认，你有足够的测试去验证你的重构是否引入了问题。

为了保证可以验证拆分过程是否引入了 BUG，需要把测试覆盖率提高到80% 以上，越高越好。

要做到：

- **点赞收藏这些服务的代码本身覆盖率有 80% 以上。**
- **使用了点赞收藏这些服务的代码覆盖率有 80% 以上。**

同时在补充完测试之后，

- **我们还要进一步检查代码，确保核心路径没有遗漏。**
- **梳理业务，确保没有关键业务场景遗漏。**

### 模块化

#### 梳理重构点

![image-20250121174038051](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211740069.png)

#### 执行拆分

![image-20250121174059555](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211741851.png)

#### 解决数据库初始化的问题

![image-20250121174125224](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202501211741153.png)

#### 挪动消费者

**在使用了 Kafka 的形态中，增加阅读数是通过订阅Kafka 的消息来实现的，所以对应的代码，我们也要**
**挪动过来。**

注意这一个部分，我们是直接复制出来一份，因为我们同样需要事件的定义。

而后删除原本和生产者有关的代码，只保留
InteractiveReadEventConsumer 相关代码，调整对应的 wire 代码。

PS：**模块化的目标就是，它不再依赖原本webook/internal 里面的任何代码**，所以不能说抽取出来一个公共的 ReadEvent，而后大家都引用它

#### 重新生成 wire 的文件

有两个地方：

- 集成测试的 wire 文件。
- main 函数启动程序的 wire 文件。

#### 模块依赖化

接下来，我们要做的就是将整个模块的代码挪动到一个新的代码仓库。

### 微服务化

#### Go 中的微服务框架

![image-20250207095022925](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502070950185.png)

#### 环境准备

![image-20250207095537195](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502070955302.png)

#### API 管理

![image-20250207095625860](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502070956971.png)

##### 目录结构

我们采用第一种方案来管理的 API，也就是在 webook下面创建一个 api 的目录，**里面放置 Protobuf 定义**。

- api 下面，理论上来说还可以放别的定义，比如说 Thrift 定义、Swagger 定义等。
- **proto 下面按照服务拆分目录**，并且有一个 gen 目录来放代码。
- **具体的服务里面，加上版本号（你也可以省略）作为目录**。

<img src="https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502071002584.png" alt="image-20250207100208677" style="zoom: 67%;" />

##### Interactive 的定义

```protobuf
syntax = "proto3";
package intr.v1;
option go_package = "webook/api/proto/gen/intr;intrv1";

service InteractiveService {
  rpc IncrReadCnt(IncrReadCntRequest) returns (IncrReadCntResponse);
  rpc Like(LikeRequest) returns (LikeResponse);
  rpc CancelLike(CancelLikeRequest) returns (CancelLikeResponse);
  rpc Collect(CollectRequest) returns (CollectResponse);
  rpc Get(GetRequest) returns (GetResponse);
  rpc GetByIds(GetByIdsRequest) returns (GetByIdsResponse);
}
```

##### 使用 buf 来简化 Protobuf 的管理难度

直接使用 protoc 有几个难点：

- 插件难管理。
- protoc 命令本身也不好记，使用起来每次都要输入一大堆的参数。
- 各种文件目录，package 定位简直要命。

所以直接使用 protoc 是比较难的。因此我们可以考虑使用 buf 来管理。

打开 https://buf.build/docs/installation，选择安装方式。

##### 使用 buf 来编译

在项目的顶级目录之下，有一个 buf.gen.yaml 文件，里面定义了 buf 怎么帮我们管理和编译 protobuf。

```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: "webook/api/proto/gen"
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: webook/api/proto/gen
    opt: paths=source_relative

  - plugin: buf.build/grpc/go
    out: webook/api/proto/gen
    opt: paths=source_relative
```



- **go_package_prefix: 避免了每次写 go package 都要写老长一段的问题。**

- **plugins: 这是比较关键的配置，这里指定了两个插件，并且在两个插件里面分别指定了 out 和 opt。**
  - Go 语言插件
  - gRPC 插件

最终在顶级目录下运行我封装好的 make grpc 就可以，又或者直接执行 buf generate webook/api/proto。

#### 实现 gRPC 接口

在 API 定义出来之后，我们需要做得就是实现这些gRPC 接口。

**注意，我们的实现都是放到了一个 gRPC 的包里面。从地位上来说 gRPC、 Job 和 Web 都是同一个级别的，是”服务对外表现形式“：**

- 对 Job 来说，就是我以定时任务的形式暴露出去。
- 对于 Web 来说，就是我以 HTTP 接口的形式暴露出去。
- 对于 gRPC 来说，就是我以 RPC 接口的形式暴露出去。

#### 调整集成测试：wire

现在需要调整集成测试。原本的集成测试是在一个统一的 integration 包里面，**这里我们将 Interactive 的集成测试分离出来，放到 Interactive 内部的 integration 里面**。同时，将原本的集成测试调整为测试 gRPC 的 Server。

> 所有的测试用例都要调整到判定 gRPC 返回的错误。

#### 启动服务：配置文件

配置文件也很简单，就是多了一个 gRPC 的启动端口，启动的时候要注意设置正确的 working directory 和 配置文件。

![image-20250208152024785](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502081520976.png)

#### 改造客户端

![image-20250208155127141](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502081551066.png)

##### 本地调用和 gRPC 调用并行方案

![image-20250208155158394](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202502081551420.png)