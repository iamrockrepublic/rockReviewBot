package main

import (
	"context"
	"os"
	"os/signal"
	"rock_review/app/bot_server"
	"rock_review/util/persist"
	"rock_review/util/xlogger"
	"syscall"
)

func initBotSvc() *bot_server.ReviewBotSvc {
	db := persist.MustNewMysqlClient("iamrock:pwd_159jkl@tcp(localhost:3306)/rock_review?charset=utf8mb4").Unsafe()

	userSessionRepo := bot_server.NewUserSessionRepo(db)
	userSessionMgr := bot_server.NewUserSessionMgr(userSessionRepo)
	reviewRepo := bot_server.NewReviewRepo(db)
	botSvc := bot_server.NewReviewBotSvc("7131845071:AAFk2z4SVHpswj3ZAnC9LY7-UJBcOuM6qC4", userSessionMgr, reviewRepo)

	return botSvc
}

func main() {
	botSvc := initBotSvc()
	ctx, cancelF := context.WithCancel(context.Background())

	err := botSvc.Init()
	if err != nil {
		xlogger.FatalF(ctx, "run bot serviced failed: %v", err)
	}
	botSvc.Run(ctx)

	xlogger.InfoF(ctx, "bot service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	<-sigChan
	xlogger.InfoF(ctx, "bot service terminating")

	cancelF()
}
