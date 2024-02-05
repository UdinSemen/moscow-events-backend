package models

import "time"

type Event struct {
	Id          string    `db:"id"`
	UrlImg      string    `db:"url_img"`
	Label       string    `db:"label"`
	Description string    `db:"description"`
	Date        time.Time `db:"date"`
	Price       string    `db:"price"`
	UrlBuy      string    `db:"url_buy"`
	IsFavorite  bool      `db:"is_favorite"`
}
