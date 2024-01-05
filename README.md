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

