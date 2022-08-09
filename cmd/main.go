package main

import (
	ordermatcher "order-matching/pkg/order-matcher"
)

func main() {
	exchange := ordermatcher.NewExchange("BTC", "USDT")

	bidderUUID, _ := exchange.AddActor(1000, 1000)
	askerUUID, _ := exchange.AddActor(1000, 1000)

	// for i := 0; i < 2; i++ {
	// 	exchange.BidOrder(bidderUUID, 1, 1000)
	// 	exchange.AskOrder(askerUUID, 1, 1000)
	// }

	exchange.BidOrder(bidderUUID, 6, 1000) // Buy
	exchange.BidOrder(bidderUUID, 1, 1000) // Buy
	exchange.AskOrder(askerUUID, 2, 1000)  // Sell
	exchange.AskOrder(askerUUID, 2, 1000)  // Sell

	// fmt.Printf("%v\n", exchange.RestingSellOrders[askOrderUUID].String())
	// fmt.Printf("%v\n", exchange.RestingBuyOrders[bidOrderUUID].String())

	exchange.MatchRestingOrders(ordermatcher.FIFOMatch)
}
