package bot_server

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"path/filepath"
	"rock_review/util/oss"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type Review struct {
	TgUserId      int64          `db:"tg_user_id"`
	ReviewContent *ReviewContent `db:"review_content"`
}

type ReviewContent struct {
	Text      string   `json:"text"`
	MediaUrls []string `json:"media_urls"`
}

func (f *ReviewContent) Value() (driver.Value, error) {
	if f == nil {
		return []byte{}, nil
	}

	bs, err := json.Marshal(f)
	return bs, err
}

func (f *ReviewContent) Scan(src any) error {
	var bs []byte
	switch convertedV := src.(type) {
	case string:
		bs = []byte(convertedV)
	case []byte:
		bs = convertedV
	default:
		return fmt.Errorf("type err")
	}

	return json.Unmarshal(bs, f)
}

type ReviewRepo struct {
	db *sqlx.DB
}

func NewReviewRepo(db *sqlx.DB) *ReviewRepo {
	return &ReviewRepo{db: db}
}

func (repo *ReviewRepo) StoreReview(ctx context.Context, userId int64, content ReviewContent) error {
	var (
		err     error
		newUrls []string
	)

	for _, mediaUrl := range content.MediaUrls {
		newUrl, err := oss.Upload(ctx, oss.BucketRockReview, uuid.NewV4().String()+filepath.Ext(mediaUrl), mediaUrl)
		if err != nil {
			return err
		}
		newUrls = append(newUrls, newUrl)
	}

	content.MediaUrls = newUrls

	review := &Review{
		TgUserId:      userId,
		ReviewContent: &content,
	}
	_, err = repo.db.NamedExecContext(ctx, "insert into review (tg_user_id, review_content) values (:tg_user_id, :review_content)", review)
	return err
}
