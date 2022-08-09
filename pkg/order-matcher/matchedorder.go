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

func MatchOrders(sellOrder, buyOrder *Order, filledBuyVolume, filledSellVolume int64) (*MatchedOrder, string) {
	if sellOrder == nil || buyOrder == nil {
		return nil, NoOrdersExhausted
	}

	if sellOrder.SlippedPrice < buyOrder.SlippedPrice {
		return nil, NoOrdersExhausted
	}

	agreedPrice := int64((sellOrder.SlippedPrice + buyOrder.SlippedPrice) / 2)
	realBuyerVolume := buyOrder.Volume - filledBuyVolume
	realSellerVolume := sellOrder.Volume - filledSellVolume
	volumeDifference := realBuyerVolume - realSellerVolume

	if volumeDifference == 0 {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, realBuyerVolume)
		return matchedOrder, BothOrdersExhausted
	} else if volumeDifference > 0 {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, realSellerVolume)
		return matchedOrder, SellOrderExhausted
	} else {
		matchedOrder := newMatchedOrder(sellOrder, buyOrder, agreedPrice, realBuyerVolume)
		return matchedOrder, BuyOrderExhausted
	}

}

func (m MatchedOrder) String() string {
	return fmt.Sprintf("%v %v @ %v %v \t %q -> %q", m.AgreedVolume, m.SellOrder.AssetType, m.AgreedPrice, m.SellOrder.BaseType, m.SellOrder.ActorUUID, m.BuyOrder.ActorUUID)
}
