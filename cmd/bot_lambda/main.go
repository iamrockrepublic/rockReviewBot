package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"rock_review/app/bot_server"
	"rock_review/util/goutil"
	"rock_review/util/persist"
	"rock_review/util/xlogger"

	"github.com/aliyun/fc-runtime-go-sdk/fc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

type HTTPTriggerEvent struct {
	Version         *string           `json:"version"`
	RawPath         *string           `json:"rawPath"`
	Headers         map[string]string `json:"headers"`
	QueryParameters map[string]string `json:"queryParameters"`
	Body            *string           `json:"body"`
	IsBase64Encoded *bool             `json:"isBase64Encoded"`
	RequestContext  *struct {
		AccountId    string `json:"accountId"`
		DomainName   string `json:"domainName"`
		DomainPrefix string `json:"domainPrefix"`
		RequestId    string `json:"requestId"`
		Time         string `json:"time"`
		TimeEpoch    string `json:"timeEpoch"`
		Http         struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			Protocol  string `json:"protocol"`
			SourceIp  string `json:"sourceIp"`
			UserAgent string `json:"userAgent"`
		} `json:"http"`
	} `json:"requestContext"`
}

// HTTPTriggerResponse HTTP Trigger Response struct
type HTTPTriggerResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers,omitempty"`
	IsBase64Encoded bool              `json:"isBase64Encoded,omitempty"`
	Body            string            `json:"body"`
}

func NewHTTPTriggerResponse(statusCode int) *HTTPTriggerResponse {
	return &HTTPTriggerResponse{StatusCode: statusCode}
}

func (h *HTTPTriggerResponse) String() string {
	jsonBytes, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

func (h *HTTPTriggerResponse) WithStatusCode(statusCode int) *HTTPTriggerResponse {
	h.StatusCode = statusCode
	return h
}

func (h *HTTPTriggerResponse) WithHeaders(headers map[string]string) *HTTPTriggerResponse {
	h.Headers = headers
	return h
}

func (h *HTTPTriggerResponse) WithIsBase64Encoded(isBase64Encoded bool) *HTTPTriggerResponse {
	h.IsBase64Encoded = isBase64Encoded
	return h
}

func (h *HTTPTriggerResponse) WithBody(body string) *HTTPTriggerResponse {
	h.Body = body
	return h
}

type httpHandler func(ctx context.Context, event HTTPTriggerEvent) (*HTTPTriggerResponse, error)

func botSvcHandler(botSvc *bot_server.ReviewBotSvc) httpHandler {
	return func(ctx context.Context, event HTTPTriggerEvent) (*HTTPTriggerResponse, error) {
		return botSvcHandle(ctx, botSvc, event)
	}
}

func botSvcHandle(ctx context.Context, botSvc *bot_server.ReviewBotSvc, event HTTPTriggerEvent) (*HTTPTriggerResponse, error) {
	if event.Body == nil {
		return NewHTTPTriggerResponse(http.StatusBadRequest).WithBody("body is nil"), nil
	}

	var (
		update tgbotapi.Update
		err    error
	)

	err = json.Unmarshal([]byte(*event.Body), &update)
	if err != nil {
		return NewHTTPTriggerResponse(http.StatusBadRequest).WithBody(err.Error()), nil
	}

	err = botSvc.HandleUpdate(ctx, update)
	if err != nil {
		return NewHTTPTriggerResponse(http.StatusBadRequest).WithBody(err.Error()), nil
	}

	return NewHTTPTriggerResponse(http.StatusOK), nil
}

func debugMiddleware(handler httpHandler) httpHandler {
	return func(ctx context.Context, event HTTPTriggerEvent) (*HTTPTriggerResponse, error) {
		var (
			resp *HTTPTriggerResponse
			err  error
		)

		panicErr := goutil.SafeDo(context.TODO(), func() {
			resp, err = handler(ctx, event)
		})
		if panicErr != nil {
			resp = NewHTTPTriggerResponse(http.StatusInternalServerError).WithBody(panicErr.Error())
		}

		logRequest(event, resp, err, panicErr)
		return resp, err
	}
}

func logRequest(req HTTPTriggerEvent, resp *HTTPTriggerResponse, err, panicErr error) {
	type logInfo struct {
		Body      string `db:"body"`
		RetStatus int    `db:"ret_status"`
		RetBody   string `db:"ret_body"`
		Err       string `db:"err"`
		PanicErr  string `db:"panic_err"`
	}

	l := &logInfo{}
	if req.Body != nil {
		l.Body = *req.Body
	}
	if resp != nil {
		l.RetStatus = resp.StatusCode
		l.RetBody = resp.Body
	}
	if err != nil {
		l.Err = err.Error()
	}
	if panicErr != nil {
		l.PanicErr = panicErr.Error()
	}

	xlogger.InfoF(context.TODO(), goutil.JsonString(l))
}

func initBotSvc() (*bot_server.ReviewBotSvc, *sqlx.DB) {
	//dbUri := "iamrock:pwd_159jkl@tcp(localhost:3306)/rock_review?charset=utf8mb4"
	dbUri := "iamrock:pwd_159jkl@tcp(rm-cn-27a3newvh0002rbo.rwlb.rds.aliyuncs.com:3306)/rock_review?charset=utf8mb4"
	reviewDb := persist.MustNewMysqlClient(dbUri).Unsafe()
	userSessionRepo := bot_server.NewUserSessionRepo(reviewDb)
	userSessionMgr := bot_server.NewUserSessionMgr(userSessionRepo)
	reviewRepo := bot_server.NewReviewRepo(reviewDb)
	botSvc := bot_server.NewReviewBotSvc("7131845071:AAFk2z4SVHpswj3ZAnC9LY7-UJBcOuM6qC4", userSessionMgr, reviewRepo)

	return botSvc, reviewDb
}

func main() {
	botSvc, db := initBotSvc()
	xlogger.Logger = dbLogger{db: db}

	err := botSvc.Init()
	if err != nil {
		xlogger.FatalF(context.Background(), "run bot serviced failed: %v", err)
	}

	xlogger.InfoF(context.TODO(), "lambda service started")

	fc.Start(debugMiddleware(botSvcHandler(botSvc)))
}

var _ xlogger.ILogger = dbLogger{}

type dbLogger struct {
	db *sqlx.DB
}

func (d dbLogger) InfoF(ctx context.Context, format string, args ...interface{}) {
	d.log(ctx, "info", format, args...)
}

func (d dbLogger) PanicF(ctx context.Context, format string, args ...interface{}) {
	d.log(ctx, "panic", format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (d dbLogger) ErrorF(ctx context.Context, format string, args ...interface{}) {
	d.log(ctx, "error", format, args...)
}

func (d dbLogger) WarnF(ctx context.Context, format string, args ...interface{}) {
	d.log(ctx, "warn", format, args...)
}

func (d dbLogger) FatalF(ctx context.Context, format string, args ...interface{}) {
	d.log(ctx, "fatal", format, args...)
	os.Exit(1)
}

func (d dbLogger) log(ctx context.Context, level string, format string, args ...interface{}) {
	type logInfo struct {
		Level   string `db:"level"`
		Content string `db:"content"`
	}
	l := &logInfo{
		Level:   level,
		Content: fmt.Sprintf(format, args...),
	}
	_, _ = d.db.NamedExec("insert into access_log (`level`, `content`) values (:level, :content)", l)
}
