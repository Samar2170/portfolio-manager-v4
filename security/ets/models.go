package ets

import (
	"time"

	"github.com/samar2170/portfolio-manager-v4/pkg/db"
	"gorm.io/gorm"
)

func init() {
	db.DB.AutoMigrate(&ETS{})
	db.DB.AutoMigrate(&ETSPriceHistory{})
}

type ETS struct {
	*gorm.Model
	ID               int
	Symbol           string `gorm:"index"`
	Name             string
	SecurityCode     string `gorm:"uniqueIndex"`
	Category         string
	PriceToBeUpdated bool
}

type ETSPriceHistory struct {
	*gorm.Model
	ID    uint
	ETS   ETS
	ETSID uint
	Price float64
	Date  time.Time
}

func (e *ETS) create() error {
	err := db.DB.Create(&e).Error
	return err
}

func (e *ETS) getOrCreate() (ETS, error) {
	err := db.DB.FirstOrCreate(&e, ETS{Symbol: e.Symbol}).Error
	return *e, err
}

func (e *ETS) getLatestDate() (time.Time, error) {
	var eph ETSPriceHistory
	err := db.DB.Order("date desc").First(&eph, "ets_id = ?", e.ID).Error
	return eph.Date, err
}

func (e *ETSPriceHistory) create() error {
	err := db.DB.Create(&e).Error
	return err
}

func GetETSBySymbol(symbol string) (ETS, error) {
	var e ETS
	err := db.DB.First(&e, "symbol = ?", symbol).Error
	return e, err
}
