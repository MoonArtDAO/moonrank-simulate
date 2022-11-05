package main

import (
	"errors"
	"github.com/Jeffail/gabs/v2"
	"log"
	"regexp"
	"strconv"
)

const DEBUG = false

func Atoi(s string) int {
	r, _ := strconv.Atoi(s)
	return r
}

var (
	IDRe = regexp.MustCompile("#(\\d+)")
)

func IDFromName(name string) string {
	matches := IDRe.FindStringSubmatch(name)
	if len(matches) == 2 {
		return matches[1]
	}

	return ""
}

func NormalizeJSON(in []byte) ([]byte, error) {
	jsonParsed, err := gabs.ParseJSON(in)
	if err != nil {
		log.Printf("NormalizeJSON: failed with err %v\n", err)
		return []byte{}, err
	}

	modified := false
	if fv, ok := jsonParsed.Path("seller_fee_basis_points").Data().(float64); !ok {
		if strValue, ok := jsonParsed.Path("seller_fee_basis_points").Data().(string); ok {
			log.Printf("NormalizeJSON: seller_fee_basis_points as string: %+v\n", strValue)
			jsonParsed.Set(Atoi(strValue), "seller_fee_basis_points")
			modified = true
		}
	} else {
		jsonParsed.Set(int(fv), "seller_fee_basis_points")
		modified = true
	}

	if sv, ok := jsonParsed.Path("collection").Data().(string); ok {
		if DEBUG {
			log.Printf("NormalizeJSON: collection as string: %+v\n", sv)
		}
		obj := gabs.New()
		obj.Set(sv, "name")
		obj.Set(sv, "family")
		jsonParsed.Set(obj, "collection")
		modified = true
	}

	if innerFile, ok := jsonParsed.Path("properties.files").Data().([]interface{}); !ok {
		if innerFile, ok := jsonParsed.Path("properties.files").Data().(map[string]interface{}); ok {
			if DEBUG {
				log.Printf("NormalizeJSON: properties.files as obj: %+v\n", innerFile)
			}
			jsonParsed.ArrayP("properties.files")
			jsonParsed.ArrayAppend(innerFile, "properties", "files")
			modified = true
		}
	} else {
		if len(innerFile) > 0 {
			if _, ok := innerFile[0].(string); ok {
				if DEBUG {
					log.Printf("NormalizeJSON: properties.files as array: %+v\n", innerFile)
				}
				jsonParsed.ArrayP("properties.files")

				for _, value := range innerFile {
					obj := gabs.New()
					obj.Set("image", "type")
					obj.Set(value, "uri")
					jsonParsed.ArrayAppend(obj, "properties", "files")
					modified = true
				}
			}
		}
	}

	if creators, ok := jsonParsed.Path("properties.creators").Data().([]interface{}); ok {
		jsonParsed.ArrayP("properties.creators")
		for _, c := range creators {
			creator, ok := c.(map[string]interface{})
			if !ok {
				log.Printf("NormalizeJSON: creators array is borked\n")
				return []byte{}, errors.New("creators array is borked")
			}

			obj := gabs.New()
			for key, value := range creator {
				if key == "share" {
					if strValue, ok := value.(string); ok {
						obj.Set(Atoi(strValue), key)
					} else {

						obj.Set(value, key)
					}
				} else {
					obj.Set(value, key)
				}
			}
			jsonParsed.ArrayAppend(obj, "properties", "creators")
			modified = true
		}
	}

	if attributes, ok := jsonParsed.Path("attributes").Data().(map[string]interface{}); ok {
		if DEBUG {
			log.Printf("NormalizeJSON: attributes as obj: %+v\n", attributes)
		}
		jsonParsed.Array("attributes")
		for key, value := range attributes {
			obj := gabs.New()
			obj.Set(key, "trait_type")
			obj.Set(value, "value")
			jsonParsed.ArrayAppend(obj, "attributes")
			modified = true
		}
	}

	if collection, ok := jsonParsed.Path("collection").Data().([]interface{}); ok {
		jsonParsed.Set(collection[0], "collection")
		modified = true
	}

	if !modified {
		return in, nil
	} else {
		if DEBUG {
			log.Printf("NormalizeJSON: modified JSON: %+v\n", jsonParsed)
		}
		return []byte(jsonParsed.String()), nil
	}
}

func StringInSlice(target string, values []string) bool {
	for _, v := range values {
		if target == v {
			return true
		}
	}

	return false
}
