package bot_server

import (
	"rock_review/util/goutil"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func Test_UnitTest_Msg_Template(t *testing.T) {
	templates := []complexText{
		{newPlainText("hello")},
		{newPlainText("hello"), newPlainText(" world")},
		{newEntityText("/hello", entityTypeBotCommand)},
		{newEntityText("/hello", entityTypeBotCommand), newEntityText("/world", entityTypeBotCommand)},
		{newPlainText("send "), newEntityText("/hello", entityTypeBotCommand), newPlainText(" to world")},
	}

	results := []struct {
		fullText string
		Entities []tgbotapi.MessageEntity
	}{
		{
			fullText: "hello",
			Entities: nil,
		},
		{
			fullText: "hello world",
			Entities: nil,
		},
		{
			fullText: "/hello",
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   entityTypeBotCommand,
					Offset: 0,
					Length: 6,
				},
			},
		},
		{
			fullText: "/hello/world",
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   entityTypeBotCommand,
					Offset: 0,
					Length: 6,
				},
				{
					Type:   entityTypeBotCommand,
					Offset: 6,
					Length: 6,
				},
			},
		},
		{
			fullText: "send /hello to world",
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   entityTypeBotCommand,
					Offset: 5,
					Length: 6,
				},
			},
		},
	}

	for i := range templates {
		msg := templates[i].buildMsg(1)
		assert.Equal(t, results[i].fullText, msg.Text, i)
		assert.Equal(t, goutil.JsonString(results[i].Entities), goutil.JsonString(msg.Entities))
	}
}
