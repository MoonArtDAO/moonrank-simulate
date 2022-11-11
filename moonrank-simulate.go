package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
)

type FilteredAttributes struct {
	Value []string
}

func (f *FilteredAttributes) String() string {
	return ""
}

func (f *FilteredAttributes) Set(s string) error {
	f.Value = append(f.Value, s)
	return nil
}

var filteredAttributes = FilteredAttributes{}

var InputDir string
var OutFile string
var DupeOutFile string

func InitFlags() {
	flag.StringVar(&InputDir, "i", "", "input directory containing *.json files")
	flag.StringVar(&OutFile, "o", "", "rarity output as json file")
	flag.StringVar(&DupeOutFile, "d", "", "duplicate output as json file")
	flag.Var(&filteredAttributes, "filtered", "filtered attributes, can be specified N times")

	flag.Parse()

	if InputDir == "" {
		log.Fatal("no input dir provided")
	}

	if OutFile == "" {
		log.Fatal("no out file provided")
	}
}

func Process(fns []string) {
	log.Printf("Process: input directory has %v metadata files.\n", len(fns))

	mints := []Mint{}
	for _, f := range fns {
		md, err := MetadataFromFile(f)
		if err != nil {
			log.Fatal(err)
		}

		mints = append(mints, MintFromMetadata(f, *md))
	}

	log.Printf("Process: %v metadata objects found.\n", len(mints))
	sortedNfts, duplicates := RankRarity(mints, filteredAttributes.Value)
	encoded, err := json.Marshal(&sortedNfts)
	if err != nil {
		log.Fatal(err)
	}

	ioutil.WriteFile(OutFile, encoded, 0644)
	log.Printf("Process: wrote rarity data to `%v`.\n", OutFile)

	if DupeOutFile != "" && len(duplicates) > 0 {
		log.Printf("%v metadata duplicates found.\n", len(duplicates))
		encoded, err := json.Marshal(&duplicates)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(DupeOutFile, encoded, 0644)
		log.Printf("Process: wrote duplicates data to `%v`.\n", DupeOutFile)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)
	InitFlags()
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Scanning `%v` for JSON files.\n", InputDir)

	fns, err := filepath.Glob(filepath.Join(InputDir, "*.json"))
	if err == nil {
		Process(fns)
	} else {
		log.Fatal(err)
	}
}
