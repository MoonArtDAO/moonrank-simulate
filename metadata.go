package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type MetadataFile struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

type MetadataAttribute struct {
	TraitType  string `json:"trait_type"`
	Value      string `json:"value"`
	Slot       int    `json:"slot"`
	Injected   bool   `json:"injected,omitempty"`
	PreMapping string `json:"pre_mapping,omitempty"`
}

type MetadataCollection struct {
	Name   string `json:"name"`
	Family string `json:"family"`
}

type Metadata struct {
	Name        string              `json:"name"`
	Symbol      string              `json:"symbol"`
	Description string              `json:"description"`
	Image       string              `json:"image"`
	ExternalURL string              `json:"external_url"`
	Attributes  []MetadataAttribute `json:"attributes"`
	Collection  *MetadataCollection `json:"collection,omitempty"`
	Properties  struct {
		Files    []MetadataFile `json:"files"`
		Category *string        `json:"category,omitempty"`
		Creators *[]struct {
			Address string `json:"address"`
			Share   int    `json:"share"`
		} `json:"creators,omitempty"`
	} `json:"properties"`
	SellerFeeBasisPoints int    `json:"seller_fee_basis_points"`
	UpdateAuthority      string `json:"update_authority,omitempty"`
}

func MetadataFromFile(f string) (md *Metadata, err error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		log.Printf("MetadataFromFile(%v): ReadFile err %v\n", f, err)
		return nil, err
	}

	b, err = NormalizeJSON(b)
	if err != nil {
		log.Printf("MetadataFromFile(%v): NormalizeJSON err %v\n", f, err)
		return nil, err
	}

	md = new(Metadata)
	jerr := json.Unmarshal(b, md)
	if jerr != nil {
		log.Printf("MetadataFromFile(%v) Unmarshal err %v\n", f, jerr)
		return nil, jerr
	}

	return md, nil
}
