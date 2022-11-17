package main

import (
	"fmt"
	"github.com/maruel/natural"
	"log"
	"sort"
	"strings"
	"time"
)

type RarityTraits map[string]map[string][]string

func AggregateRarity(nfts []Mint) RarityTraits {
	traitMap := make(RarityTraits)

	for _, mi := range nfts {
		n := mi.Metadata

		for _, v := range n.Attributes {
			t := strings.TrimSpace(v.TraitType)
			x := strings.TrimSpace(v.Value)

			_, ok := traitMap[t]
			if !ok {
				traitMap[t] = make(map[string][]string)
			}

			traitMap[t][x] = append(traitMap[t][x], mi.ID)
		}
	}

	return traitMap
}

func AggregateRarityByRankInfo(nfts []Mint) RarityTraits {
	traitMap := make(RarityTraits)

	for _, mi := range nfts {
		for _, v := range mi.RankExplain {
			t := strings.TrimSpace(v.Attribute)
			x := strings.TrimSpace(v.Value)

			_, ok := traitMap[t]
			if !ok {
				traitMap[t] = make(map[string][]string)
			}

			traitMap[t][x] = append(traitMap[t][x], mi.ID)
		}
	}

	return traitMap
}

func MintRankInfoCmp(x, y MintRankInfo) bool {
	if x.TimesSeen < y.TimesSeen {
		return true
	}

	if x.TimesSeen == y.TimesSeen {
		if x.Attribute == y.Attribute {
			return natural.Less(x.Value, y.Value)
		}

		return natural.Less(x.Attribute, y.Attribute)
	}

	return false
}

func MintFilteredRankInfoCmp(x, y MintFilteredRankInfo) bool {
	if x.TimesSeen < y.TimesSeen {
		return true
	}

	if x.TimesSeen == y.TimesSeen {
		if x.Attribute == y.Attribute {
			return natural.Less(x.Value, y.Value)
		}

		return natural.Less(x.Attribute, y.Attribute)
	}

	return false
}

func SortRankExplain(ri []MintRankInfo) []MintRankInfo {
	sort.SliceStable(ri, func(i, j int) bool {
		return MintRankInfoCmp(ri[i], ri[j])
	})

	return ri
}

func SortFilteredRankExplain(ri []MintFilteredRankInfo) []MintFilteredRankInfo {
	n := []MintFilteredRankInfo{}
	for _, i := range ri {
		if !i.Frivolous {
			n = append(n, i)
		}
	}

	sort.SliceStable(n, func(i, j int) bool {
		return MintFilteredRankInfoCmp(n[i], n[j])
	})

	return n
}

func FindMaximumShape(mi []Mint) map[string]int {
	shape := make(map[string]int)

	for _, x := range mi {
		n := x.Metadata
		traitCount := make(map[string]int)

		for _, t := range n.Attributes {
			v, _ := traitCount[t.TraitType]
			traitCount[t.TraitType] = v + 1
		}

		for k, v := range traitCount {
			maxCount, ok := shape[k]
			if !ok {
				shape[k] = 1
			} else if v > maxCount {
				shape[k] = v
			}
		}
	}

	log.Printf("FindMaximumShape: shape %+v\n", shape)
	return shape
}

func CreateSlots(mi []Mint, shape map[string]int) []Mint {
	for mintIdx, x := range mi {
		n := x.Metadata

		for k, c := range shape {
			found := 0
			for i, t := range n.Attributes {
				if k == t.TraitType {
					n.Attributes[i].Slot = found
					found = found + 1
				}
			}

			if found != c {
				for i := found; i < c; i++ {
					log.Printf("CreateSlots: creating empty slot %v:`%v` for %v (%v)\n",
						i, k, n.Name, x.ID)
					n.Attributes = append(n.Attributes,
						MetadataAttribute{Slot: i, TraitType: k, Value: "", Injected: true})
				}
			}
		}

		mi[mintIdx].Metadata = n
	}

	return mi
}

func MapTraits(mi []Mint, mapTraits map[string]string) []Mint {
	if len(mapTraits) == 0 {
		return mi
	}

	for mintIdx, x := range mi {
		n := x.Metadata

		for k, v := range mapTraits {
			for aIdx, t := range n.Attributes {
				if k == t.TraitType {
					log.Printf("MapTraits: mapping `%v` to `%v` for %v\n",
						t.TraitType, v, n.Name)
					n.Attributes[aIdx].TraitType = v
					break
				}
			}
		}

		mi[mintIdx].Metadata = n
	}

	return mi
}

func MapValues(mi []Mint, mapValues map[string]map[string]string) []Mint {
	if len(mapValues) == 0 {
		return mi
	}

	for mintIdx, x := range mi {
		n := x.Metadata

		for k, v := range mapValues {
			for aIdx, t := range n.Attributes {
				if k == t.TraitType {
					replace, ok := v[t.Value]
					if ok {
						log.Printf("MapValues: mapping `%v` to `%v` in `%v` for %v\n",
							t.Value, replace, n.Name, x.ID)
						n.Attributes[aIdx].PreMapping = n.Attributes[aIdx].Value
						n.Attributes[aIdx].Value = replace
					}
					break
				}
			}
		}

		mi[mintIdx].Metadata = n
	}

	return mi
}

func NormalizeTraits(mi []Mint) []Mint {
	for mintIdx, x := range mi {
		n := x.Metadata
		newAttributes := []MetadataAttribute{}

		for _, t := range n.Attributes {
			if len(strings.TrimSpace(t.TraitType)) == 0 {
				continue
			}

			newAttributes = append(newAttributes, t)
		}

		n.Attributes = newAttributes
		mi[mintIdx].Metadata = n
	}

	return mi
}

var AutoFilteredAttributes = []string{"sequence", "generation", "Sequence", "Collection", "Rarity Rank"}

func RankRarity(nfts []Mint, inFilteredAttributes []string) (sortedNfts []Mint, outDupes [][]string) {
	ms := time.Now()

	frivolousAttributes := []string{}
	filteredAttributes := []string{}
	for _, fa := range inFilteredAttributes {
		filteredAttributes = append(filteredAttributes, fa)
	}

	for _, fa := range AutoFilteredAttributes {
		if !StringInSlice(fa, filteredAttributes) {
			filteredAttributes = append(filteredAttributes, fa)
		}
	}

	log.Printf("RankRarity: filteredAttributes = %+v\n", filteredAttributes)

	nfts = NormalizeTraits(nfts)

	shape := FindMaximumShape(nfts)

	nfts = CreateSlots(nfts, shape)

	traitMap := AggregateRarity(nfts)

	total := len(nfts)

	for i, mi := range nfts {
		n := mi.Metadata

		absoluteRarity := float64(1)
		filteredRarity := float64(1)
		rankExplains := []MintRankInfo{}
		filteredExplains := []MintFilteredRankInfo{}

		sort.SliceStable(n.Attributes, func(i, j int) bool {
			return strings.Compare(
				n.Attributes[i].TraitType,
				n.Attributes[j].TraitType) < 0
		})

		for _, a := range n.Attributes {
			tt := strings.TrimSpace(a.TraitType)
			tv := strings.TrimSpace(a.Value)
			timesSeen := len(traitMap[tt][tv])

			attrR := float64(timesSeen) / float64(total)
			attrScore := attrR * float64(100.)
			x := MintRankInfo{Attribute: tt,
				Value:           tv,
				TimesSeen:       timesSeen,
				TotalSeen:       total,
				ValuePercentage: attrScore}

			absoluteRarity = absoluteRarity * attrR
			if StringInSlice(tt, filteredAttributes) {
				filteredExplains = append(filteredExplains,
					MintFilteredRankInfo{
						MintRankInfo: x,
						Frivolous: StringInSlice(tt, AutoFilteredAttributes) ||
							StringInSlice(tt, frivolousAttributes),
					},
				)
				continue
			}

			filteredRarity = filteredRarity * attrR
			rankExplains = append(rankExplains, x)
		}

		nfts[i].RankExplain = SortRankExplain(rankExplains)
		nfts[i].FilteredRankExplain = SortFilteredRankExplain(filteredExplains)
		nfts[i].AbsoluteRarity = absoluteRarity
		nfts[i].FilteredRarity = filteredRarity
	}

	sort.SliceStable(nfts, func(i, j int) bool {
		if nfts[i].FilteredRarity == nfts[j].FilteredRarity {
			if strings.Compare(nfts[i].Metadata.Name, nfts[j].Metadata.Name) == 0 {
				return strings.Compare(nfts[i].ID, nfts[j].ID) < 0
			} else {
				return natural.Less(nfts[i].Metadata.Name, nfts[j].Metadata.Name)
			}
		}

		return nfts[i].FilteredRarity < nfts[j].FilteredRarity
	})

	dupesByMintAddr := make(map[string][]string)

	for i, n := range nfts {
		if i > 0 && nfts[i-1].FilteredRarity == nfts[i].FilteredRarity {
			nfts[i].Rank = nfts[i-1].Rank
		} else {
			nfts[i].Rank = i + 1
		}

		dupeKey := fmt.Sprintf("%v", n.RankExplain)
		dupesByMintAddr[dupeKey] = append(dupesByMintAddr[dupeKey], n.ID)

		log.Printf("Rank %d: %v (Rarity %.16f) [%v]\n", nfts[i].Rank, n.Metadata.Name, n.FilteredRarity, dupeKey)
	}

	totalDupes := 0

	for _, x := range dupesByMintAddr {
		if len(x) > 1 {
			totalDupes = totalDupes + len(x)
			outDupes = append(outDupes, x)
		}
	}

	log.Printf("RankRarity: took %v to rank\n", time.Now().Sub(ms))
	log.Printf("Dupes: %d\n", totalDupes)
	return nfts, outDupes
}
