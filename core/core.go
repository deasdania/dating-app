package core

import (
	"time"

	ps "github.com/deasdania/dating-app/storage/postgresql"
	"github.com/sirupsen/logrus"
)

type Core struct {
	log     *logrus.Entry
	storage *ps.Storage
	td      time.Duration
}

const randomListgenerating = 10

// NewCore will create new a Core object representation of ICore interface
func NewCore(log *logrus.Entry, storage *ps.Storage, td time.Duration) *Core {
	return &Core{
		log:     log,
		storage: storage,
		td:      td,
	}
}
