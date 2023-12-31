package pstock

import (
	"errors"
	"strconv"

	"github.com/samar2170/portfolio-manager-v4/internal/models"
	"github.com/samar2170/portfolio-manager-v4/pkg/db"
	"github.com/samar2170/portfolio-manager-v4/security/stock"
	"gorm.io/gorm"
)

type StockTrade struct {
	*gorm.Model
	ID        int
	StockID   int
	Stock     *stock.Stock
	Quantity  int
	Price     float64
	TradeType string
	TradeDate string
	Account   models.DematAccount
}

func NewStockTrade(symbol string, quantity, price, tradeDate, tradeType string) (*StockTrade, error) {
	stock, err := stock.GetStockBySymbol(symbol)
	if err != nil {
		return nil, err
	}
	quantityParsed, err := strconv.ParseInt(quantity, 10, 64)
	if err != nil {
		return nil, err
	}
	priceParsed, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return nil, err
	}
	return &StockTrade{
		StockID:   stock.ID,
		Quantity:  int(quantityParsed),
		Price:     priceParsed,
		TradeType: tradeType,
		TradeDate: tradeDate,
	}, nil
}

type StockHolding struct {
	*gorm.Model
	StockID  int
	Stock    *stock.Stock
	Quantity int
	BuyPrice float64
	Account  models.DematAccount
}

func (s *StockTrade) create() error {
	return db.DB.Create(s).Error
}
func (s *StockTrade) GetAccount() models.DematAccount {
	return s.Account
}
func (s *StockTrade) GetInvestedValue() float64 {
	return s.Price * float64(s.Quantity)
}

func (s *StockHolding) create() error {
	return db.DB.Create(s).Error
}
func (s *StockHolding) update() error {
	return db.DB.Save(s).Error
}
func (s *StockHolding) GetInvestedValue() float64 {
	return float64(s.Quantity) * s.BuyPrice
}

func RegisterStockTrade(s *StockTrade) error {
	err := s.create()
	if err != nil {
		return err
	}
	existingHolding := stockHoldingExists(s.StockID, s.Account.UserCID)
	if existingHolding {
		holding, err := getStockHolding(s.StockID, s.Account.UserCID)
		if err != nil {
			return err
		}
		if s.TradeType == "buy" {
			holding.Quantity += s.Quantity
			holding.BuyPrice = (holding.GetInvestedValue() + s.GetInvestedValue()) / (float64(holding.Quantity) + float64(s.Quantity))
		} else {
			holding.Quantity -= s.Quantity
		}
		err = holding.update()
		if err != nil {
			return err
		}
	} else {
		if s.TradeType == "sell" {
			return errors.New("cannot sell stock that you do not own")
		} else {
			holding := StockHolding{
				StockID:  s.StockID,
				Quantity: s.Quantity,
				BuyPrice: s.Price,
				Account:  s.Account,
			}
			err := holding.create()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getStockHolding(stockId int, userCID string) (StockHolding, error) {
	var stockHolding StockHolding
	dematAccounts, _ := models.GetDematAccountsByUserCID(userCID)
	dematIds := make([]int, len(dematAccounts))
	for i, account := range dematAccounts {
		dematIds[i] = account.ID
	}
	err := db.DB.Where("stock_id = ? AND account_id IN ?", stockId, dematIds).First(&stockHolding).Error
	return stockHolding, err
}

func stockHoldingExists(stockId int, userCID string) bool {
	return db.DB.Where("stock_id = ? AND account_id IN ?", stockId, userCID).First(&StockHolding{}).Error == nil
}
