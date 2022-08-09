package ordermatcher

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	SellOrderExhausted  = "Sell Order Exhausted"
	BuyOrderExhausted   = "Buy Order Exhausted"
	BothOrdersExhausted = "Both Orders Exhausted"
	NoOrdersExhausted   = "No Orders Exhausted"
)

type MatchedOrder struct {
	UUID         uuid.UUID
	SellOrder    *Order
	BuyOrder     *Order
	AgreedVolume int64
	AgreedPrice  int64
}

func newMatchedOrder(sellOrder, buyOrder *Order, price, volume int64) *MatchedOrder {
	return &MatchedOrder{
		UUID:         uuid.New(),
		SellOrder:    sellOrder,
		BuyOrder:     buyOrder,
		AgreedPrice:  price,
		AgreedVolume: volume,
	}
}

func MatchOrders(sellOrder, buyOrder *Order, lastExhaustionStatus string, lastRemainingVolume int64) (*MatchedOrder, string, int64) {
	if sellOrder == nil || buyOrder == nil {
		return nil, NoOrdersExhausted, 0
	}

	if sellOrder.Slip < buyOrder.Slip {
		return nil, NoOrdersExhausted, 0
	}

	var volumeDifference int64
	agreedPrice := int64((sellOrder.Slip + buyOrder.Slip) / 2)

	if lastExhaustionStatus == BuyOrderExhausted {
		volumeDifference = buyOrder.Volume + lastRemainingVolume - sellOrder.Volume
	} else {
		volumeDifference = buyOrder.Volume - lastRemainingVolume - sellOrder.Volume
	}

	if volumeDifference == 0 {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, sellOrder.Volume)
		return matchedOrder, BothOrdersExhausted, 0
	} else if volumeDifference > 0 {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, sellOrder.Volume)
		return matchedOrder, SellOrderExhausted, volumeDifference
	} else {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, buyOrder.Volume)
		return matchedOrder, BuyOrderExhausted, -volumeDifference
	}

}

func (m MatchedOrder) String() string {
	return fmt.Sprintf("%v %v @ %v %v \t %q -> %q", m.AgreedVolume, m.SellOrder.AssetType, m.AgreedPrice, m.SellOrder.BaseType, m.SellOrder.ActorUUID, m.BuyOrder.ActorUUID)
}
