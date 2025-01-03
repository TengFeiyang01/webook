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

![image-20240103140237448](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941098.png)

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

![image-20240104091836551](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941617.png)

解决方法： **preflight请求** ：需要在 `preflight` 请求中告诉浏览器，**我这个 `localhost:8090` 能够接收 ** `localhost:3000` **过来的请求。**

**preflight请求** 的特征：`preflight` 请求会发到同一个地址上，使用 `Options` 方法，没有请求参数

![image-20240104093431844](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941477.png)

![image-20240104093450681](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941126.png)

###### 使用 `middleware` 来解决 `CORS`

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

![image-20240105105347456](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941966.png)

`dao` 中的 `User` 模型：注意到，`dao` 里面操作的不是 `domain.User` ，而是新定义了一个类型。

这是因为：**`domain.User` 是业务概念，它不一定和数据库中表或者列完全对的上。而 `dao.User` 是直接映射到表里面的。**

那么问题就来了：**如何建表？**

![image-20240105112220350](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941280.png)

#### 密码加密

- 谁来加密？`service` 还是 `repository` 还是 `dao` ？
- 怎么加密？怎么选择一个安全的加密算法？

**PS：敏感信息应该是连日志都不能打**

###### 加密的位置：

![image-20240105113719864](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941959.png)

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

![image-20240105135814571](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941029.png)

## 用户登录

#### 实现登录功能

![image-20240105144629664](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941586.png)

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

![image-20240105162049868](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941252.png)

![image-20240105162319316](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180941240.png)

#### 使用 `Gin` 的 `Session` 插件来实现登录功能

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

##### 使用`Redis`

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

##### `Session` 参数

![image-20240108090839733](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409180942717.png)

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

## 发表文章

**发表文章接口，我们使用单元测试TDD。** 

![image-20240925143852501](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409251438435.png)

### Web层测试

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

## 维护状态

### 状态图

![image-20240926160955535](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261609704.png)

### 状态定义

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

### 状态流转

![image-20240926163028118](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409261630335.png)

![image-20240927103210140](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032442.png)

![](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032442.png)

![image-20240927103245248](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032631.png)

![image-20240927103256243](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271032290.png)

## MongoDB

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

### DAO 抽象

![image-20240927154300249](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271543317.png)

![image-20240927164342150](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271643430.png)

![image-20240927164356114](C:/Users/ytf/AppData/Roaming/Typora/typora-user-images/image-20240927164356114.png)

![image-20240927164407448](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271644387.png)

![image-20240927164419919](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409271644093.png)

![image-20240929100822929](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008070.png)

![image-20240929100836278](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008257.png)

![image-20240929100845729](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291008902.png)

![image-20240929100856769](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291009529.png)

### 发布接口重构

![image-20240929100934226](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291009535.png)

# 利用OSS来存储数据

![image-20240929101945124](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291019028.png)

## OSS入门

![image-20240929102442763](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291024702.png)

![image-20240929102812219](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291028165.png)

![image-20240929103027601](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291030767.png)

## S3 API 入门

![image-20240929103124701](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291031695.png)

![image-20240929103138352](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291031569.png)

![image-20240929104805388](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291048394.png)

![image-20240929104836033](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049403.png)

![image-20240929104854362](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049086.png)

![image-20240929104910219](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049374.png)![image-20240929104918563](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291049665.png)

## 总结

![image-20240929111614793](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291116979.png)

![image-20240929111712853](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291117883.png)

![image-20240929111545509](https://gcore.jsdelivr.net/gh/TengFeiyang01/picture@master/data/202409291115231.png)
