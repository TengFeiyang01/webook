package repository

import (
	"golang.org/x/net/context"
	"webook/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
