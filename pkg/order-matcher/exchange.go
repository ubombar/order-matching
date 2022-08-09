package ordermatcher

import (
	"errors"
	"math/rand"
	"sort"
	"time"

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
	rand.Seed(time.Now().UnixNano())

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

func (e *Exchange) MatchRestingOrders(algorithm string) *MatchResult {
	defer e.resetRestingOrders()

	switch algorithm {
	case FIFOMatch:
		return e.matchRestingOrdersFIFOMatch()
	}

	return nil
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

func (e *Exchange) matchRestingOrdersFIFOMatch() *MatchResult {
	result := NewMatchResult()
	sellOrders := make([]*Order, e.RestingSellOrderCount)
	buyOrders := make([]*Order, e.RestingBuyOrderCount)
	filledBuyVolume := int64(0)
	filledSellVolume := int64(0)
	sellerIndex := 0
	buyerIndex := 0

	for _, sellOrder := range e.RestingSellOrders {
		sellOrders[sellOrder.RestingOrderIndex] = sellOrder
	}

	for _, buyOrder := range e.RestingBuyOrders {
		buyOrders[buyOrder.RestingOrderIndex] = buyOrder
	}

	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price == sellOrders[j].Price {
			return sellOrders[i].RestingOrderIndex < sellOrders[j].RestingOrderIndex
		}
		return sellOrders[i].SlippedPrice < sellOrders[j].SlippedPrice
	})

	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price == buyOrders[j].Price {
			return buyOrders[i].RestingOrderIndex < buyOrders[j].RestingOrderIndex
		}
		return buyOrders[i].SlippedPrice > buyOrders[j].SlippedPrice
	})

	for sellerIndex < e.RestingSellOrderCount && buyerIndex < e.RestingBuyOrderCount {
		currentSellOrder := sellOrders[sellerIndex]
		currentBuyOrder := buyOrders[buyerIndex]
		matchedOrder, exhaustionStatus := MatchOrders(currentSellOrder, currentBuyOrder, filledBuyVolume, filledSellVolume)

		if exhaustionStatus == NoOrdersExhausted {
			break
		} else if exhaustionStatus == BothOrdersExhausted {
			result.TotalVolume += matchedOrder.AgreedVolume
			currentBuyOrder.Status = OrderStatusFullyFilled
			currentSellOrder.Status = OrderStatusFullyFilled
			filledSellVolume = 0
			filledBuyVolume = 0
			sellerIndex += 1
			buyerIndex += 1
		} else if exhaustionStatus == SellOrderExhausted {
			result.TotalVolume += matchedOrder.AgreedVolume
			currentSellOrder.Status = OrderStatusFullyFilled
			currentBuyOrder.Status = OrderStatusPartiallyFilled
			filledSellVolume = 0
			filledBuyVolume += matchedOrder.AgreedVolume
			sellerIndex += 1
		} else {
			result.TotalVolume += matchedOrder.AgreedVolume
			currentSellOrder.Status = OrderStatusPartiallyFilled
			currentBuyOrder.Status = OrderStatusFullyFilled
			filledSellVolume += matchedOrder.AgreedVolume
			filledBuyVolume = 0
			buyerIndex += 1
		}

		if result.HighestPrice < matchedOrder.AgreedPrice {
			result.HighestPrice = matchedOrder.AgreedPrice
		}

		if result.LowestPrice > matchedOrder.AgreedPrice {
			result.LowestPrice = matchedOrder.AgreedPrice
		}

		result.Matches = append(result.Matches, matchedOrder)
	}

	return result
}
