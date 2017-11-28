package ordered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestMarshalOrderedMap(t *testing.T) {
	om := NewOrderedMap()
	om.Set("a", 34)
	om.Set("b", []int{3, 4, 5})
	b, err := json.Marshal(om)
	if err != nil {
		t.Fatalf("Marshal OrderedMap: %v", err)
	}
	// fmt.Printf("%q\n", b)
	const expected = "{\"a\":34,\"b\":[3,4,5]}"
	if !bytes.Equal(b, []byte(expected)) {
		t.Errorf("Marshal OrderedMap: %q not equal to expected %q", b, expected)
	}
}

func ExampleDecodeOrderedMap() {
	const jsonStream = `{
  "country"     : "United States",
  "countryCode" : "US",
  "region"      : "CA",
  "regionName"  : "California",
  "city"        : "Mountain View",
  "zip"         : "94043",
  "lat"         : 37.4192,
  "lon"         : -122.0574,
  "timezone"    : "America/Los_Angeles",
  "isp"         : "Google Cloud",
  "org"         : "Google Cloud",
  "as"          : "AS15169 Google Inc.",
  "mobile"      : true,
  "proxy"       : false,
  "query"       : "35.192.25.53"
}`

	// compare with if using a regular generic map, the unmarshalled result is a map with unpredictable order of keys
	var m map[string]interface{}
	err := json.Unmarshal([]byte(jsonStream), &m)
	if err != nil {
		fmt.Println("error:", err)
	}
	for key := range m {
		// fmt.Printf("%-12s: %v\n", key, m[key])
		_ = key
	}

	// use the OrderedMap to Unmarshal from JSON object
	var om *OrderedMap = NewOrderedMap()
	err = json.Unmarshal([]byte(jsonStream), om)
	if err != nil {
		fmt.Println("error:", err)
	}
	for pair := range om.Entries() {
		fmt.Printf("%-12s: %v\n", pair.Key, pair.Value)
	}

	// Output:
	// country     : United States
	// countryCode : US
	// region      : CA
	// regionName  : California
	// city        : Mountain View
	// zip         : 94043
	// lat         : 37.4192
	// lon         : -122.0574
	// timezone    : America/Los_Angeles
	// isp         : Google Cloud
	// org         : Google Cloud
	// as          : AS15169 Google Inc.
	// mobile      : true
	// proxy       : false
	// query       : 35.192.25.53
}

func TestUnmarshalOrderedMapFromInvalid(t *testing.T) {
	om := NewOrderedMap()
	err := json.Unmarshal([]byte("[]"), om)
	if err == nil {
		t.Fatal("Unmarshal OrderedMap: expecting error")
	}
}

func TestUnmarshalOrderedMap(t *testing.T) {
	var (
		data = []byte(`{"as":"AS15169 Google Inc.","city":"Mountain View","country":"United States","countryCode":"US","isp":"Google Cloud","lat":37.4192,"lon":-122.0574,"org":"Google Cloud","query":"35.192.25.53","region":"CA","regionName":"California","status":"success","timezone":"America/Los_Angeles","zip":"94043"}`)
		obj  = &OrderedMap{
			m: map[string]interface{}{
				"as":          "AS15169 Google Inc.",
				"city":        "Mountain View",
				"country":     "United States",
				"countryCode": "US",
				"isp":         "Google Cloud",
				"lat":         37.4192,
				"lon":         -122.0574,
				"org":         "Google Cloud",
				"query":       "35.192.25.53",
				"region":      "CA",
				"regionName":  "California",
				"status":      "success",
				"timezone":    "America/Los_Angeles",
				"zip":         "94043",
			},
			keys: []string{
				"as", "city", "country", "countryCode", "isp",
				"lat", "lon", "org", "query", "region", "regionName",
				"status", "timezone", "zip",
			},
		}
	)

	om := NewOrderedMap()
	err := json.Unmarshal(data, om)
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}

	// fix number type for deepequal test
	for _, key := range []string{"lat", "lon"} {
		num, _ := om.Get(key)
		numf, _ := num.(json.Number).Float64()
		om.Set(key, numf)
	}

	b, err := json.MarshalIndent(om, "", "  ")
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}
	const expected = `{
  "as": "AS15169 Google Inc.",
  "city": "Mountain View",
  "country": "United States",
  "countryCode": "US",
  "isp": "Google Cloud",
  "lat": 37.4192,
  "lon": -122.0574,
  "org": "Google Cloud",
  "query": "35.192.25.53",
  "region": "CA",
  "regionName": "California",
  "status": "success",
  "timezone": "America/Los_Angeles",
  "zip": "94043"
}`
	if !bytes.Equal(b, []byte(expected)) {
		t.Fatalf("Unmarshal OrderedMap marshal indent from %#v not equal to expected: %q\n", om, expected)
	}

	if !reflect.DeepEqual(om, obj) {
		t.Fatalf("Unmarshal OrderedMap not deeply equal: %#v %#v", om, obj)
	}
}

func TestUnmarshalNestedOrderedMap(t *testing.T) {
	var (
		data = []byte(`{"a": true, "b": [3, 4, { "b": "3", "d": [] }]}`)
		obj  = &OrderedMap{
			m: map[string]interface{}{
				"a": true,
				"b": []interface{}{3, 4, &OrderedMap{
					m: map[string]interface{}{
						"b": "3",
						"d": []interface{}{},
					},
					keys: []string{"b", "d"},
				}},
			},
			keys: []string{"a", "b"},
		}
	)
	om := NewOrderedMap()
	err := json.Unmarshal(data, om)
	if err != nil {
		t.Fatalf("Unmarshal OrderedMap: %v", err)
	}

	// b, err := json.MarshalIndent(om, "", "  ")
	// fmt.Println(om, string(b), err, obj)

	// fix number type for deepequal test
	ele, _ := om.Get("b")
	elearr := ele.([]interface{})
	for i, v := range elearr {
		if num, ok := v.(json.Number); ok {
			numi, _ := num.Int64()
			elearr[i] = int(numi)
		}
	}

	if !reflect.DeepEqual(om, obj) {
		t.Fatalf("Unmarshal OrderedMap not deeply equal: %#v expected %#v", om, obj)
	}
}
