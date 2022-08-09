package ordermatcher

import (
	"fmt"
	"math"
)

type MatchResult struct {
	Matches      []*MatchedOrder
	TotalVolume  int64
	HighestPrice int64
	LowestPrice  int64
}

func NewMatchResult() *MatchResult {
	return &MatchResult{
		Matches:      make([]*MatchedOrder, 0, 1),
		TotalVolume:  0,
		HighestPrice: math.MinInt64,
		LowestPrice:  math.MaxInt64,
	}
}

func (r MatchResult) String() string {
	if len(r.Matches) == 0 {
		return "0 trx."
	}
	assetType := r.Matches[0].SellOrder.AssetType
	baseType := r.Matches[0].SellOrder.BaseType
	return fmt.Sprintf("%v trxs, total of %v %v. high %v %v low %v %v", len(r.Matches), r.TotalVolume, assetType, r.HighestPrice, baseType, r.LowestPrice, baseType)
}
