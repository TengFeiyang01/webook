package failover

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"github.com/TengFeiyang01/webook/webook/internal/service/sms"
	smssvcmocks "github.com/TengFeiyang01/webook/webook/internal/service/sms/mocks"
)

func TestFailoverSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) []sms.Service

		wantErr error
	}{
		{
			name: "一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smssvcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0}
			},

			wantErr: nil,
		},
		{
			name: "重试成功",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smssvcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("无法发送"))
				svc1 := smssvcmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			wantErr: nil,
		},
		{
			name: "重试最终失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := smssvcmocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("无法发送"))
				svc1 := smssvcmocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("无法发送"))
				return []sms.Service{svc0, svc1}
			},
			wantErr: errors.New("发送失败，所有服务商都尝试过了"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewFailoverSMSService(tc.mock(ctrl))
			err := svc.Send(context.Background(), "mytpl", []string{"123"}, "132333")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
