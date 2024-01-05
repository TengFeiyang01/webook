package grpc

import "context"

type UserService interface {
	// @path /users/:id
	// @method GET
	// @header
	// @authorization
	GetById(ctx context.Context, id int64) (User, error)
}

// GO 使用反射来注入一个 GetById 的值
type UserServiceV1 struct {
	GetById func(
		ctx context.Context,
		id int64) (User, error) `method:"get" path:"/user/:id"`
}

// JAVA 动态代理生成一个实现
// GO 代码生成
type UserServiceHttpImpl struct {
}

func (u *UserServiceHttpImpl) GetById(ctx context.Context, id int64) (User, error) {
	// 组装 HTTP 请求
	//TODO implement me
	panic("implement me")
}

//type User struct {
//}

//func JavaStyle() {
//	future := asyncInvoke(req)
//	// 干别的事情
//
//	// 拿 future 的结果
//	resp := future.Get()
//}

//func GoStyle() {
//	var res Result
//	go func() {
//		res = Invoke(req)
//	}()
//
//	// 干别的事情
//
//	// 操作 res
//}

//func JavaStyle() {
//	Invoke(req, callback)
//}

//func GoStyle() {
//	go func() {
//		res := Invoke(req)
//		handle(res)
//	}()
//}

//func UserGen() {
//	var user User
//	user.Contacts.isUser_Contacts()
//
//	ee, ok := user.Contacts.(*User_Email)
//	if ok {
//		// 邮箱
//	} else {
//		// 就是手机号码
//	}
//}
