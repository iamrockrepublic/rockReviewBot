package oss

import "context"

const (
	BucketRockReview = "ROCK_REVIEW"
)

func Upload(ctx context.Context, targetBucket string, targetPath string, sourceUrl string) (newUrl string, err error) {
	return sourceUrl, nil
}
