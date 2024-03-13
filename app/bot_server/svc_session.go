package bot_server

import (
	"context"
	"rock_review/util/xlogger"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	sessionStateInit    sessionState = 0
	sessionStateComment sessionState = 1
)

type sessionState = int

type UserSession struct {
	userSessionData
	chatId int64

	lastActiveUnix int64 // lastActiveUnix is a meta property, accessing it requires a meta lock
	updateCh       chan tgbotapi.Update
	once           sync.Once
	cancelF        context.CancelFunc

	reviewBot       IReviewBotSvc
	reviewRepo      IReviewRepo
	userSessionRepo ISessionRepo
}

func NewUserSession(data userSessionData, chatId int64, botSvc IReviewBotSvc, reviewRepo IReviewRepo, sessionRepo ISessionRepo) *UserSession {
	userSession := &UserSession{
		userSessionData: data,
		chatId:          chatId,
		lastActiveUnix:  time.Now().Unix(),
		updateCh:        make(chan tgbotapi.Update, 10),
		reviewBot:       botSvc,
		reviewRepo:      reviewRepo,
		userSessionRepo: sessionRepo,
	}

	return userSession
}

func (session *UserSession) Run(ctx context.Context) {
	session.once.Do(func() {
		newCtx, cancelF := context.WithCancel(ctx)
		session.cancelF = cancelF

		xlogger.InfoF(ctx, "session start for user: %d", session.UserId)
		for {
			select {
			case <-newCtx.Done():
				return
			case update, ok := <-session.updateCh:
				if !ok {
					return
				}
				session.handleUpdate(newCtx, update)
			}
		}
	})
}

func (session *UserSession) Close() {
	if session.cancelF != nil {
		session.cancelF()
	}
}

func (session *UserSession) Save(ctx context.Context) error {
	return session.userSessionRepo.SetUserSessionData(ctx, session.userSessionData)
}

func (session *UserSession) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		session.handleUserMsg(ctx, update.Message)
	}
}

func (session *UserSession) handleUserMsg(ctx context.Context, message *tgbotapi.Message) {
	var (
		err error
	)
	xlogger.InfoF(ctx, "%s wrote %s", message.From.FirstName, message.Text)

	if len(message.Text) > 0 && message.Text[0] == '/' {
		err = session.handleCommand(ctx, message)
		if err != nil {
			xlogger.ErrorF(ctx, "handle user command %s fail: %v", message.Text, err)
			_, _ = session.reviewBot.Send(ctx, errRetryTemplate.buildMsg(session.chatId))
		}
		return
	}

	if message.Contact != nil {
		if message.Contact.UserID != message.From.ID {
			_, _ = session.reviewBot.Send(ctx, phoneBindUseOwnContactTemplate.buildMsg(session.chatId))
			return
		}
		session.PhoneNumber = message.Contact.PhoneNumber
		err = session.Save(ctx)
		if err != nil {
			xlogger.ErrorF(ctx, "save phone number fail:%v", err)
			_, _ = session.reviewBot.Send(ctx, errRetryTemplate.buildMsg(session.chatId))
			return
		}
		successMsg := phoneBindSuccessTemplate.buildMsg(session.chatId)
		successMsg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
		_, _ = session.reviewBot.Send(ctx, successMsg)
		return
	}

	if session.State == sessionStateComment {
		err = session.handleUserComment(ctx, message)
		if err != nil {
			xlogger.ErrorF(ctx, "handle user comment fail: %v", err)
			_, _ = session.reviewBot.Send(ctx, errRetryTemplate.buildMsg(session.chatId))
		}
		return
	}

	session.sendHelpMsg(ctx)
}

func (session *UserSession) sendHelpMsg(ctx context.Context) {
	msg := helpMsgTemplate.buildMsg(session.chatId)
	if len(session.PhoneNumber) == 0 { // todo do not always pop if user refuse to provide phone number
		msg = helpMsgRequestPhoneTemplate.buildMsg(session.chatId)
		msg.ReplyMarkup = requestPhoneMarkup
	}
	_, _ = session.reviewBot.Send(ctx, msg)
}

func (session *UserSession) handleUserComment(ctx context.Context, message *tgbotapi.Message) (err error) {
	if len(message.Text) == 0 && len(message.Photo) == 0 && message.Video == nil && message.Audio == nil && message.Voice == nil {
		_, _ = session.reviewBot.Send(ctx, sendValidCommentTemplate.buildMsg(session.chatId))
		return
	}

	var (
		mediaUrls []string
		fileIds   []string
		text      string
	)

	if len(message.Photo) > 0 {
		fileIds = append(fileIds, message.Photo[len(message.Photo)-1].FileID)
	}
	if message.Video != nil {
		fileIds = append(fileIds, message.Video.FileID)
	}
	if message.Audio != nil {
		fileIds = append(fileIds, message.Audio.FileID)
	}
	if message.Voice != nil {
		fileIds = append(fileIds, message.Voice.FileID)
	}

	for _, fileId := range fileIds {
		mediaUrl, err := session.reviewBot.GetFileUrl(ctx, fileId)
		if err != nil {
			return err
		}
		mediaUrls = append(mediaUrls, mediaUrl)
	}

	text = message.Text
	if len(message.Caption) != 0 {
		text = message.Caption
	}

	err = session.reviewRepo.StoreReview(ctx, session.UserId, ReviewContent{
		Text:      text,
		MediaUrls: mediaUrls,
	})
	if err != nil {
		return err
	}
	_, _ = session.reviewBot.Send(ctx, resumeCommentTemplate.buildMsg(session.chatId))
	return
}

func (session *UserSession) handleCommand(ctx context.Context, message *tgbotapi.Message) (err error) {

	switch message.Text {
	case "/start":
		session.sendHelpMsg(ctx)
	case "/comment":
		session.State = sessionStateComment
		err = session.Save(ctx)
		if err != nil {
			return err
		}
		_, _ = session.reviewBot.Send(ctx, startCommentTemplate.buildMsg(session.chatId))
	//case "/comment_transaction":
	//	_, _ = session.reviewBot.Send(ctx, complexText{newPlainText("you used /comment_transaction")}.buildMsg(session.chatId))
	case "/finish":
		session.State = sessionStateInit
		err = session.Save(ctx)
		if err != nil {
			return err
		}
		_, _ = session.reviewBot.Send(ctx, finishCommentTemplate.buildMsg(session.chatId))
	default:
		_, _ = session.reviewBot.Send(ctx, unknownCommandTemplate.buildMsg(session.chatId))
	}
	return nil
}
