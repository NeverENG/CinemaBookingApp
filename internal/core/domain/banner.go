package domain

import "time"

// Banner 首页运营位：只存外部图片 URL，不落文件。
type Banner struct {
	ID        int64
	Title     string
	ImageURL  string
	Sort      int
	Enabled   bool
	CreatedAt time.Time
}

func (b *Banner) Validate() error {
	if b.Title == "" || b.ImageURL == "" || b.Sort < 0 {
		return ErrBannerInvalid
	}
	return nil
}
