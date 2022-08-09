package ordermatcher

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

const (
	DefaultRestingOrderLength = 1000
	DefaultSlippagePercent    = float64(0.01)

	FIFOMatch = "FIFO Match"
)

type Exchange struct {
	Actors                map[uuid.UUID]*Actor
	RestingBuyOrders      map[uuid.UUID]*Order
	RestingSellOrders     map[uuid.UUID]*Order
	RestingBuyOrderCount  int
	RestingSellOrderCount int
	AssetType             string
	BaseType              string
}

func NewExchange(assetType, baseType string) *Exchange {
	// rand.Seed(time.Now().UnixNano())
	rand.Seed(123)

	exchange := Exchange{
		Actors:                make(map[uuid.UUID]*Actor),
		RestingBuyOrders:      make(map[uuid.UUID]*Order),
		RestingSellOrders:     make(map[uuid.UUID]*Order),
		RestingBuyOrderCount:  0,
		RestingSellOrderCount: 0,
		AssetType:             assetType,
		BaseType:              baseType,
	}

	return &exchange
}

func (e *Exchange) AddActor(assetVolume, baseVolume int64) (uuid.UUID, error) {
	actor := NewActor(assetVolume, baseVolume)

	if _, ok := e.Actors[actor.UUID]; ok {
		return uuid.Nil, errors.New("actor already exists")
	}

	e.Actors[actor.UUID] = actor

	return actor.UUID, nil
}

func (e *Exchange) AskOrder(actorUUID uuid.UUID, volume, price int64) (uuid.UUID, error) {
	slip := int64(float64(price) * (1.0 + DefaultSlippagePercent))
	order, err := NewOrder(actorUUID, SellOrder, price, volume, slip, e.AssetType, e.BaseType, e.RestingSellOrderCount)

	e.RestingSellOrders[order.UUID] = order
	e.RestingSellOrderCount += 1

	return order.UUID, err
}

func (e *Exchange) BidOrder(actorUUID uuid.UUID, volume, price int64) (uuid.UUID, error) {
	slip := int64(float64(price) * (1.0 - DefaultSlippagePercent))
	order, err := NewOrder(actorUUID, BuyOrder, price, volume, slip, e.AssetType, e.BaseType, e.RestingBuyOrderCount)

	e.RestingBuyOrders[order.UUID] = order
	e.RestingBuyOrderCount += 1

	return order.UUID, err
}

func (e *Exchange) MatchRestingOrders(algorithm string) {
	defer e.resetRestingOrders()

	switch algorithm {
	case FIFOMatch:
		e.matchRestingOrdersFIFOMatch()
	}
}

func (e *Exchange) resetRestingOrders() {
	e.RestingBuyOrderCount = 0
	e.RestingSellOrderCount = 0

	for u := range e.RestingBuyOrders {
		delete(e.RestingBuyOrders, u)
	}

	for u := range e.RestingSellOrders {
		delete(e.RestingSellOrders, u)
	}
}

func (e *Exchange) matchRestingOrdersFIFOMatch() {
	sellOrders := make([]*Order, e.RestingSellOrderCount)
	buyOrders := make([]*Order, e.RestingBuyOrderCount)

	for _, sellOrder := range e.RestingSellOrders {
		sellOrders[sellOrder.RestingOrder] = sellOrder
	}

	for _, buyOrder := range e.RestingBuyOrders {
		buyOrders[buyOrder.RestingOrder] = buyOrder
	}

	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].RestingOrder == sellOrders[j].RestingOrder {
			return sellOrders[i].RestingOrder < sellOrders[j].RestingOrder
		}
		return sellOrders[i].Slip < sellOrders[j].Slip
	})

	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].RestingOrder == buyOrders[j].RestingOrder {
			return buyOrders[i].RestingOrder < buyOrders[j].RestingOrder
		}
		return buyOrders[i].Slip > buyOrders[j].Slip
	})

	var lastExhaustionStatus string
	var lastRemainingVolume int64
	var matchedOrderList []*MatchedOrder
	sellerIndex := 0
	buyerIndex := 0

	for sellerIndex < e.RestingSellOrderCount && buyerIndex < e.RestingBuyOrderCount {
		currentSellOrder := sellOrders[sellerIndex]
		currentBuyOrder := buyOrders[buyerIndex]
		matchedOrder, exhaustionStatus, remainingVolume := MatchOrders(currentSellOrder, currentBuyOrder, lastExhaustionStatus, lastRemainingVolume)

		fmt.Printf("%v %v %v\n", matchedOrder, exhaustionStatus, remainingVolume)

		if exhaustionStatus == NoOrdersExhausted {
			break
		}

		lastRemainingVolume = remainingVolume
		matchedOrderList = append(matchedOrderList, matchedOrder)

		if exhaustionStatus == BothOrdersExhausted {
			currentSellOrder.Status = OrderStatusFullyFilled
			currentBuyOrder.Status = OrderStatusFullyFilled
			sellerIndex += 1
			buyerIndex += 1
		} else if exhaustionStatus == SellOrderExhausted {
			currentSellOrder.Status = OrderStatusFullyFilled
			currentBuyOrder.Status = OrderStatusPartiallyFilled
			sellerIndex += 1
		} else {
			currentSellOrder.Status = OrderStatusPartiallyFilled
			currentBuyOrder.Status = OrderStatusFullyFilled
			buyerIndex += 1
		}
	}

	fmt.Println("")
	for _, e := range matchedOrderList {
		fmt.Printf("%v\n", e)
	}

	for _, value := range e.RestingBuyOrders {
		fmt.Printf("%v\n", value)
	}

	for _, value := range e.RestingSellOrders {
		fmt.Printf("%v\n", value)
	}

}
