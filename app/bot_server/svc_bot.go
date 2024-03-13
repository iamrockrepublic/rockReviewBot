package bot_server

import (
	"context"
	"fmt"
	"rock_review/util/goutil"
	"rock_review/util/xlogger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ReviewBotSvc struct {
	userSessionMgr *UserSessionMgr
	reviewRepo     *ReviewRepo

	botToken string
	botApi   *tgbotapi.BotAPI
}

func NewReviewBotSvc(botToken string, userSessionMgr *UserSessionMgr, reviewRepo *ReviewRepo) *ReviewBotSvc {
	return &ReviewBotSvc{
		botToken:       botToken,
		userSessionMgr: userSessionMgr,
		reviewRepo:     reviewRepo,
	}
}

func (bot *ReviewBotSvc) Run(ctx context.Context) {
	goutil.SafeGo(ctx, func() {
		bot.receiveUpdates(ctx)
	})
	return
}

func (bot *ReviewBotSvc) Init() error {
	var err error
	bot.botApi, err = tgbotapi.NewBotAPI(bot.botToken)
	return err
}

func (bot *ReviewBotSvc) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.SentFrom() == nil || update.FromChat() == nil {
		return fmt.Errorf("sent_from or from_chat cannot be nil")
	}
	// todo is chat_id unchanged for a certain user_id
	userSession := bot.userSessionMgr.GetCurrentUserSession(ctx, update.SentFrom().ID, update.FromChat().ID, bot)

	userSession.handleUpdate(ctx, update)
	return nil
}

func (bot *ReviewBotSvc) receiveUpdates(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.botApi.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			bot.botApi.StopReceivingUpdates()
		case update := <-updates:
			if update.SentFrom() == nil || update.FromChat() == nil {
				continue
			}
			// todo is chat_id unchanged for a certain user_id
			userSession := bot.userSessionMgr.GetCurrentUserSession(ctx, update.SentFrom().ID, update.FromChat().ID, bot)

			select {
			// todo gracefully wait all session done, then close
			case userSession.updateCh <- update:
			default:
				xlogger.ErrorF(ctx, "userSession channel full, ignoring update, tg_user_id: %d", update.SentFrom().ID)
			}

		}
	}
}

func (bot *ReviewBotSvc) Send(ctx context.Context, c tgbotapi.Chattable) (tgbotapi.Message, error) {
	msg, err := bot.botApi.Send(c)
	if err != nil {
		xlogger.ErrorF(ctx, "send msg failed: %v", err)
	}
	return msg, err
}

func (bot *ReviewBotSvc) GetFileUrl(ctx context.Context, fileId string) (string, error) {
	file, err := bot.botApi.GetFile(tgbotapi.FileConfig{FileID: fileId})
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.botToken, file.FilePath)
	return url, nil
}
