package db

import (
	"github.com/jinzhu/gorm"

	"check-multiple-mic-connection/config"
)

type DB interface {
	Connect(config *config.Config) DB
	GetConn() *gorm.DB
	Close() error
}
