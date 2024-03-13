package bot_server

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type userSessionData struct {
	UserId      int64        `db:"tg_user_id"`
	State       sessionState `db:"state"`
	PhoneNumber string       `db:"phone_number"`
}

type UserSessionRepo struct {
	//redisCli *redis.Client
	db *sqlx.DB
}

func NewUserSessionRepo(db *sqlx.DB) *UserSessionRepo {
	repo := &UserSessionRepo{
		db: db,
	}
	return repo
}

func (repo *UserSessionRepo) GetUserSessionData(ctx context.Context, userId int64) (userSessionData, error) {
	var (
		sessionData userSessionData
		err         error
	)

	err = repo.db.GetContext(ctx, &sessionData, "select * from review_user_session where tg_user_id = ?", userId)
	if err == sql.ErrNoRows {
		sessionData = repo.newInitSessionData(userId)
		err = repo.SetUserSessionData(ctx, sessionData)
		if err != nil {
			return userSessionData{}, err
		}
		return sessionData, nil
	}

	return sessionData, err
}

func (repo *UserSessionRepo) newInitSessionData(userId int64) userSessionData {
	return userSessionData{
		UserId: userId,
		State:  sessionStateInit,
	}
}

func (repo *UserSessionRepo) SetUserSessionData(ctx context.Context, sessionData userSessionData) error {
	var (
		err error
	)

	_, err = repo.db.ExecContext(ctx,
		"insert into review_user_session (tg_user_id, phone_number, `state`) values (?,?,?) "+
			"on duplicate key update phone_number = ?, `state` = ?",
		sessionData.UserId, sessionData.PhoneNumber, sessionData.State, sessionData.PhoneNumber, sessionData.State)
	return err
}
