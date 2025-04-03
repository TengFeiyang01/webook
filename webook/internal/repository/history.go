package repository

import (
	"golang.org/x/net/context"
	"github.com/TengFeiyang01/webook/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
