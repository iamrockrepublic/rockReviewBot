package bot_server

import (
	"context"
	"fmt"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/golang/mock/gomock"
)

const (
	testChatId      int64 = 8989
	testUserId      int64 = 3678
	testPhoneNumber       = "86478901"
)

func Test_UnitTest_SessionHandleUpdate(t *testing.T) {

	var (
		ctx = context.Background()
	)

	t.Run("input: /start", func(t *testing.T) {
		t.Run("no phone number", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/start").message)

			// set up expectation
			msg := helpMsgRequestPhoneTemplate.buildMsg(testChatId)
			msg.ReplyMarkup = requestPhoneMarkup
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, msg)

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("have phone number", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId:      testUserId,
				State:       sessionStateInit,
				PhoneNumber: testPhoneNumber,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/start").message)

			// set up expectation
			msg := helpMsgTemplate.buildMsg(testChatId)
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, msg)

			// do test
			session.handleUpdate(ctx, u.update)
		})

	})

	t.Run("input: /comment", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/comment").message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			})
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, startCommentTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("fail", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/comment").message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			}).Return(fmt.Errorf("save fail"))
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, errRetryTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})
	})

	t.Run("input: /finish", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/finish").message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			})
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, finishCommentTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("fail", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/finish").message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}).Return(fmt.Errorf("save fail"))
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, errRetryTemplate.buildMsg(testChatId))
			// do test
			session.handleUpdate(ctx, u.update)
		})
	})

	t.Run("input: contact", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetContact(&tgbotapi.Contact{
				PhoneNumber: testPhoneNumber,
				UserID:      testUserId,
			}).message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId:      testUserId,
				State:       sessionStateInit,
				PhoneNumber: testPhoneNumber,
			})
			successMsg := phoneBindSuccessTemplate.buildMsg(testChatId)
			successMsg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true}
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, successMsg)

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("fail: wrong user_id", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetContact(&tgbotapi.Contact{
				PhoneNumber: testPhoneNumber,
				UserID:      testUserId - 1,
			}).message)

			// set up expectation
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, phoneBindUseOwnContactTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("fail: save err", func(t *testing.T) {
			// init dependency
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateInit,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetContact(&tgbotapi.Contact{
				PhoneNumber: testPhoneNumber,
				UserID:      testUserId,
			}).message)

			// set up expectation
			dep.sessionRepoCtrl.EXPECT().SetUserSessionData(ctx, userSessionData{
				UserId:      testUserId,
				State:       sessionStateInit,
				PhoneNumber: testPhoneNumber,
			}).Return(fmt.Errorf("save fail"))
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, errRetryTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})
	})

	t.Run("input: other", func(t *testing.T) {
		t.Run("init state", func(t *testing.T) {
			t.Run("no phone number", func(t *testing.T) {
				dep := mockDependency(t)
				session := NewUserSession(userSessionData{
					UserId: testUserId,
					State:  sessionStateInit,
				}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

				// set up parameters
				u := newMockUpdate()
				u.SetMessage(newMockMessage().SetText("hi").message)

				// set up expectation
				msg := helpMsgRequestPhoneTemplate.buildMsg(testChatId)
				msg.ReplyMarkup = requestPhoneMarkup
				dep.reviewBotSvcCtrl.EXPECT().Send(ctx, msg)

				// do test
				session.handleUpdate(ctx, u.update)
			})

			t.Run("have phone number", func(t *testing.T) {
				dep := mockDependency(t)
				session := NewUserSession(userSessionData{
					UserId:      testUserId,
					State:       sessionStateInit,
					PhoneNumber: testPhoneNumber,
				}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

				// set up parameters
				u := newMockUpdate()
				u.SetMessage(newMockMessage().SetText("hi").message)

				// set up expectation
				dep.reviewBotSvcCtrl.EXPECT().Send(ctx, helpMsgTemplate.buildMsg(testChatId))

				// do test
				session.handleUpdate(ctx, u.update)
			})
		})
		t.Run("comment state", func(t *testing.T) {
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("hi").message)

			// set up expectation
			dep.reviewRepoCtrl.EXPECT().StoreReview(ctx, testUserId, ReviewContent{Text: "hi"})
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, resumeCommentTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})

		t.Run("unknown command", func(t *testing.T) {
			dep := mockDependency(t)
			session := NewUserSession(userSessionData{
				UserId: testUserId,
				State:  sessionStateComment,
			}, testChatId, dep.reviewBotSvcCtrl, dep.reviewRepoCtrl, dep.sessionRepoCtrl)

			// set up parameters
			u := newMockUpdate()
			u.SetMessage(newMockMessage().SetText("/some_cmd").message)

			// set up expectation
			dep.reviewBotSvcCtrl.EXPECT().Send(ctx, unknownCommandTemplate.buildMsg(testChatId))

			// do test
			session.handleUpdate(ctx, u.update)
		})
	})
}

type sessionDependency struct {
	reviewBotSvcCtrl *MockIReviewBotSvc
	reviewRepoCtrl   *MockIReviewRepo
	sessionRepoCtrl  *MockISessionRepo
}

func mockDependency(tb testing.TB) *sessionDependency {
	ctrl := gomock.NewController(tb)
	reviewBotSvcMock := NewMockIReviewBotSvc(ctrl)
	reviewRepoMock := NewMockIReviewRepo(ctrl)
	sessionRepoMock := NewMockISessionRepo(ctrl)

	dep := &sessionDependency{
		reviewBotSvcCtrl: reviewBotSvcMock,
		reviewRepoCtrl:   reviewRepoMock,
		sessionRepoCtrl:  sessionRepoMock,
	}
	return dep
}

type mockUpdate struct {
	update tgbotapi.Update
}

func (mu *mockUpdate) SetMessage(message *tgbotapi.Message) {
	mu.update.Message = message
}

type mockMessage struct {
	message *tgbotapi.Message
}

func (mm *mockMessage) SetText(text string) *mockMessage {
	mm.message.Text = text
	return mm
}

func (mm *mockMessage) SetContact(contact *tgbotapi.Contact) *mockMessage {
	mm.message.Contact = contact
	return mm
}

func newMockUpdate() *mockUpdate {
	u := tgbotapi.Update{}
	return &mockUpdate{update: u}
}

func newMockMessage() *mockMessage {
	m := &tgbotapi.Message{
		From: &tgbotapi.User{ID: testUserId},
		Chat: &tgbotapi.Chat{ID: testChatId},
	}
	return &mockMessage{message: m}
}
