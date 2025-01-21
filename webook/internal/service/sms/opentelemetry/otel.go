package opentelemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"webook/webook/internal/service/sms"
)

type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("webook/webook/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

func (s Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send_"+biz,
		// 我是一个调用短信服务商的客户端
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	span.AddEvent("发送短信")
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
