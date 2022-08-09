package main

import (
	"fmt"
	"math/rand"
	ordermatcher "order-matching/pkg/order-matcher"
)

func main() {
	exchange := ordermatcher.NewExchange("BMB", "mUSDT")

	bidderUUID, _ := exchange.AddActor(1000, 1000)
	askerUUID, _ := exchange.AddActor(1000, 1000)

	for i := 0; i < 200; i++ {
		exchange.BidOrder(bidderUUID, 5+rand.Int63n(3), 1010+rand.Int63n(10))
		exchange.AskOrder(askerUUID, 5+rand.Int63n(3), 1000+rand.Int63n(10))
	}

	matchedOrders := exchange.MatchRestingOrders(ordermatcher.FIFOMatch)

	fmt.Printf("%v\n", matchedOrders)

	for _, e := range matchedOrders.Matches {
		fmt.Printf("%v\n", e)
	}
}
