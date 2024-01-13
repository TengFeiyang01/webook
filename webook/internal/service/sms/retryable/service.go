package retryable

// Service
// 这个要小心并发问题
//type Service struct {
//	svc sms.Service
//	// 重试
//	retryCnt int
//}
//
//func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
//	err := s.svc.Send(ctx, tpl, args, numbers...)
//	for err != nil && s.retryCnt < 10 {
//		err = s.svc.Send(ctx, tpl, args, numbers...)
//		s.retryCnt++
//	}
//	return err
//}
