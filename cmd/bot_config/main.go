package main

import (
	"fmt"
	"io"
	"net/http"
	"rock_review/util/goutil"
	"strings"
)

func main() {
	botToken := "7131845071:AAFk2z4SVHpswj3ZAnC9LY7-UJBcOuM6qC4"
	method, data := methodSetWebhook()

	bodyReader := strings.NewReader(goutil.JsonString(data))

	resp, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/%s", botToken, method), "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("resp: %s", string(rawResp))
}

func methodSetWebhook() (string, map[string]any) {
	method := "setWebhook"

	data := map[string]any{
		"url": "https://review-bot-svc-renrxplzls.ap-southeast-1.fcapp.run",
	}
	return method, data
}

func methodGetWebhookInfo() (string, map[string]any) {
	method := "getWebhookInfo"

	data := map[string]any{}
	return method, data
}
