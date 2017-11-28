// Package ordered provided a type OrderedMap for use in JSON handling
// although JSON spec says the keys order of an object should not matter
// but sometimes when working with particular third-party proprietary code
// which has incorrect using the keys order, we have to maintain the object keys
// in the same order of incoming JSON object, this package is useful for these cases.
package ordered

// Refers
//  JSON and Go        https://blog.golang.org/json-and-go
//  Go-Ordered-JSON    https://github.com/virtuald/go-ordered-json
//  Python OrderedDict https://github.com/python/cpython/blob/2.7/Lib/collections.py#L38

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type m map[string]interface{}

type OrderedMap struct {
	m
	keys []string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{m: make(m)}
}

func (om *OrderedMap) Keys() []string { return om.keys }
func (om *OrderedMap) Set(key string, value interface{}) {
	if _, ok := om.m[key]; !ok {
		om.keys = append(om.keys, key)
	}
	om.m[key] = value
}

// Get value for particular key, or nil if the key does not exist
func (om *OrderedMap) Get(key string) (value interface{}, ok bool) {
	value, ok = om.m[key]
	return
}

// TODO: delete is not efficient unless implement a DoubleLL
// deletes the element with the specified key (m[key]) from the map. If there is no such element, this is a no-op.
func (om *OrderedMap) Delete(key string) {
	if _, ok := om.m[key]; ok {
		// delete from om.keys
	}
	// delete(om.m, key)
}

// Iterate all key/value pairs in the same order of object constructed
func (om *OrderedMap) Entries() <-chan struct {
	Key   string
	Value interface{}
} {
	res := make(chan struct {
		Key   string
		Value interface{}
	})
	go func() {
		for _, key := range om.keys {
			res <- struct {
				Key   string
				Value interface{}
			}{key, om.m[key]}
		}
		close(res)
	}()
	return res
}

// this implements type json.Marshaler, so can be called in json.Marshal(om)
func (om *OrderedMap) MarshalJSON() (res []byte, err error) {
	res = append(res, '{')
	for i, k := range om.keys {
		res = append(res, fmt.Sprintf("%q:", k)...)
		var b []byte
		b, err = json.Marshal(om.m[k])
		if err != nil {
			return
		}
		res = append(res, b...)
		if i < len(om.keys)-1 {
			res = append(res, ',')
		}
	}
	res = append(res, '}')
	// fmt.Printf("marshalled: %v: %#v\n", res, res)
	return
}

// this implements type json.Unmarshaler, so can be called in json.Marshal(data, om)
func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expect JSON object open with '{'")
	}

	om.parseobject(dec)

	t, err = dec.Token()
	if err != io.EOF {
		return fmt.Errorf("expect end of JSON object but got more token: %T: %v or err: %v", t, t, err)
	}

	return nil
}

func (om *OrderedMap) parseobject(dec *json.Decoder) (err error) {
	var t json.Token
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return err
		}
		om.keys = append(om.keys, key)
		om.m[key] = value
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}

func parsearray(dec *json.Decoder) (arr []interface{}, err error) {
	var t json.Token
	arr = make([]interface{}, 0)
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return
		}
		arr = append(arr, value)
	}
	t, err = dec.Token()
	if err != nil {
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		err = fmt.Errorf("expect JSON array close with ']'")
		return
	}

	return
}

func handledelim(t json.Token, dec *json.Decoder) (res interface{}, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewOrderedMap()
			err = om2.parseobject(dec)
			if err != nil {
				return
			}
			return om2, nil
		case '[':
			var value []interface{}
			value, err = parsearray(dec)
			if err != nil {
				return
			}
			return value, nil
		default:
			return nil, fmt.Errorf("Unexpected delimiter: %q", delim)
		}
	}
	return t, nil
}
