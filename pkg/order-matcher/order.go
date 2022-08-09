package ordermatcher

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

const (
	BuyOrder  = "Buy Order"
	SellOrder = "Sell Order"

	OrderStatusResting         = "Resting"
	OrderStatusPartiallyFilled = "Partially Filled"
	OrderStatusFullyFilled     = "Fully Filled"
)

type Order struct {
	UUID         uuid.UUID
	ActorUUID    uuid.UUID
	Type         string
	Price        int64
	Volume       int64
	Slip         int64
	AssetType    string
	BaseType     string
	RestingOrder int
	Status       string
}

func NewOrder(actorUUID uuid.UUID,
	orderType string,
	price,
	volume,
	slip int64,
	assetType,
	baseType string,
	restingOrder int) (*Order, error) {
	if orderType != BuyOrder && orderType != SellOrder {
		return nil, errors.New("orderType can ony be buy or sell")
	}

	if slip < 0 {
		return nil, errors.New("slip cannot be negative")
	}

	if volume < 0 {
		return nil, errors.New("volume cannot be negative")
	}

	order := Order{
		UUID:         uuid.New(),
		ActorUUID:    actorUUID,
		Type:         orderType,
		Price:        price,
		Volume:       volume,
		Slip:         slip,
		AssetType:    assetType,
		BaseType:     baseType,
		RestingOrder: restingOrder,
		Status:       OrderStatusResting,
	}

	return &order, nil
}

func (o Order) String() string {
	var askOrBid string
	var filled string
	var slip string

	if o.Slip-o.Price < 0 {
		slip = fmt.Sprintf("%v", o.Slip-o.Price)
	} else {
		slip = fmt.Sprintf("+%v", o.Slip-o.Price)
	}

	if o.Status == OrderStatusResting {
		filled = "resting"
	} else if o.Status == OrderStatusPartiallyFilled {
		filled = "partial"
	} else {
		filled = "filled"
	}

	if o.Type == BuyOrder {
		askOrBid = "BID"
	} else {
		askOrBid = "ASK"
	}

	return fmt.Sprintf("%v[%v] %v %v @ %v%v %v (%v)", askOrBid, o.RestingOrder, o.Volume, o.AssetType, o.Price, slip, o.BaseType, filled)
}
