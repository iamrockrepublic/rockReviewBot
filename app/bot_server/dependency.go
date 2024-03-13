//go:generate mockgen -source dependency.go -package bot_server -destination dependency_go_mock.go
package bot_server

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type IReviewBotSvc interface {
	Send(ctx context.Context, c tgbotapi.Chattable) (tgbotapi.Message, error)
	GetFileUrl(ctx context.Context, fileId string) (string, error)
}

type IReviewRepo interface {
	StoreReview(ctx context.Context, userId int64, content ReviewContent) error
}

type ISessionRepo interface {
	SetUserSessionData(ctx context.Context, sessionData userSessionData) error
}
