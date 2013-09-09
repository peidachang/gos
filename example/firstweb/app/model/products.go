package model

import (
	"time"
)

type ProductModel struct {
	Id        float64
	Title     string
	Price     float64
	CreatedAt time.Time
}
