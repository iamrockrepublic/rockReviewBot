package bot_server

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type entityType = string

const (
	entityTypeBotCommand entityType = "bot_command"
)

type complexText []iTextComponent

func (ct complexText) buildMsg(chatId int64) tgbotapi.MessageConfig {
	var (
		builder  = strings.Builder{}
		entities []tgbotapi.MessageEntity
	)

	for _, component := range ct {
		offset := builder.Len()
		builder.WriteString(component.Text())
		if entityText, ok := component.(*textComponentEntity); ok {
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   entityText.entityType,
				Offset: offset,
				Length: len(entityText.Text()),
			})
		}
	}

	msg := tgbotapi.NewMessage(chatId, builder.String())
	msg.Entities = entities
	return msg
}

type iTextComponent interface {
	Text() string
}

type textComponentPlainText struct {
	text string
}

func newPlainText(text string) iTextComponent {
	return &textComponentPlainText{text: text}
}

func (plainText *textComponentPlainText) Text() string {
	return plainText.text
}

func newEntityText(text string, entityType string) iTextComponent {
	return &textComponentEntity{
		text:       text,
		entityType: entityType,
	}
}

type textComponentEntity struct {
	text       string
	entityType string
}

func (entityText *textComponentEntity) Text() string {
	return entityText.text
}

var (
	helpMsgTemplate = complexText{newPlainText("Pleased to serve you.\n\n"),
		newEntityText("/comment", entityTypeBotCommand), newPlainText(" - start to comment\n"),
		//newEntityText("/comment_transaction", entityTypeBotCommand), newPlainText(" - comment on specific transaction\n"),
		newEntityText("/finish", entityTypeBotCommand), newPlainText(" - finish a comment"),
	}

	helpMsgRequestPhoneTemplate = append(helpMsgTemplate, newPlainText("\n\nTo help us serve you better, you may provide you phone number to bind your RockShop account with telegram account."))

	requestPhoneMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: [][]tgbotapi.KeyboardButton{
			{tgbotapi.NewKeyboardButtonContact("Provide phone number for better service")},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
	phoneBindSuccessTemplate       = complexText{newPlainText("Success! This telegram account is linked to your RockShop account.")}
	phoneBindUseOwnContactTemplate = complexText{newPlainText("Provide your contact to update your phone_number, not other's")}

	unknownCommandTemplate = complexText{newPlainText("Unrecognized command. Say what?")}

	startCommentTemplate  = complexText{newPlainText("Please send your comment, you can send text, image, video, audio or voice.")}
	resumeCommentTemplate = complexText{newPlainText("Your comment has been accepted, you can continue to add more, or use "),
		newEntityText("/finish", entityTypeBotCommand), newPlainText(" to finish your comment")}
	sendValidCommentTemplate = complexText{newPlainText("Please send text, image, video, audio or voice")}
	finishCommentTemplate    = complexText{newPlainText("Thanks for your reply, happy to serve you.")}

	errRetryTemplate = complexText{newPlainText("Unknown error occurred, please retry later")}
)
