package tencent

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20190711"
	"os"
	"testing"
)

func TestService_Send(t *testing.T) {
	secretID, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal("SMS_SECRET_ID not set")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		t.Fatal("SMS_SECRET_KEY not set")
	}
	fmt.Println(secretID, secretKey)
	c, err := sms.NewClient(common.NewCredential(secretID, secretKey),
		"ap-nanjing",
		profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}

	// todo: 换成自己的
	s := NewService(c, "1400933146", "f8f771e40fb513565430580b343500c8")

	testCases := []struct {
		name    string
		tplID   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:    "发送验证码",
			tplID:   "1877556",
			params:  []string{"123456"},
			numbers: []string{"13219820609"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplID, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
