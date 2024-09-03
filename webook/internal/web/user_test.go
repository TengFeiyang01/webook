package web

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"webook/webook/internal/domain"
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
		name string
	}{
		{},
	}
	//req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(`
	//{
	//	"email": "123@qq.com",
	//	"password": "123456"
	//}
	//`)))

	// 这里就可以继续使用 req 了

	//resp := httptest.NewRecorder()
	//h := NewUserHandler(nil, nil)
	t.Log("")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//ctx := &gin.Context{}
			//handler.SignUp(ctx)
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
