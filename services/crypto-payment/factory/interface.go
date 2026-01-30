package factory

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/common"
)

// SQLiteWrapper defines the behavior of a SQLite database wrapper
type SQLiteWrapper interface {
	Get(id uint64) (*common.BalanceEntry, error)
	Add() (uint64, error)
	GetAll() ([]*common.BalanceEntry, error)
	Close() error
	IsInterfaceNil() bool
}

// APIHandler defines the operations supported by the API
type APIHandler interface {
	GetConfig(c *gin.Context)
	CreateAddress(c *gin.Context)
	GetAccount(c *gin.Context)
}

// HTTPServer defines the operations supported by the HTTP server
type HTTPServer interface {
	Start() error
	GetAddress() string
	Close() error
}

// BalanceProcessor defines the operations supported by a component able to process balance changes
type BalanceProcessor interface {
	ProcessAll(ctx context.Context) error
	IsInterfaceNil() bool
}
