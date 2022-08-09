package ordermatcher

import (
	"github.com/google/uuid"
)

type Actor struct {
	UUID        uuid.UUID
	AssetVolume int64
	BaseVolume  int64
}

func NewActor(assetVolume, baseVolume int64) *Actor {
	actor := Actor{
		UUID:        uuid.New(),
		AssetVolume: assetVolume,
		BaseVolume:  baseVolume,
	}

	return &actor
}
