package main

import (
	"fmt"
	"github.com/mr-tron/base58"
	"log"
	"path/filepath"
)

type MintRankInfo struct {
	Attribute       string  `json:"attribute"`
	Value           string  `json:"value"`
	ValuePercentage float64 `json:"value_perc"`
	TimesSeen       int     `json:"times_seen"`
	TotalSeen       int     `json:"total_seen"`
}

type MintFilteredRankInfo struct {
	MintRankInfo
	Frivolous bool `json:"frivolous"`
}

type Mint struct {
	ID string `json:"mint"`

	Metadata Metadata `json:"metadata"`

	Rank           int     `json:"rank"`
	AbsoluteRarity float64 `json:"absolute_rarity"`
	FilteredRarity float64 `json:"filtered_rarity"`

	RankExplain         []MintRankInfo         `json:"rank_explain"`
	FilteredRankExplain []MintFilteredRankInfo `json:"filtered_rank_explain"`
}

func StableIdFromMetadata(f string, md Metadata) string {
	fb := filepath.Base(f)
	id := fmt.Sprintf("%v:%v", fb, md.Name)

	chk := base58.Encode([]byte(id))
	if DEBUG {
		log.Printf("input %v, output %v\n", id, chk)
	}
	return chk
}

func MintFromMetadata(f string, md Metadata) (mint Mint) {
	id := StableIdFromMetadata(f, md)

	mint.ID = id
	mint.Metadata = md

	return mint
}
