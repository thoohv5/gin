// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin/internal/json"
)

// EnableDecoderUseNumber is used to call the UseNumber method on the JSON
// Decoder instance. UseNumber causes the Decoder to unmarshal a number into an
// interface{} as a Number instead of as a float64.
var EnableDecoderUseNumber = false

// EnableDecoderDisallowUnknownFields is used to call the DisallowUnknownFields method
// on the JSON Decoder instance. DisallowUnknownFields causes the Decoder to
// return an error when the destination is a struct and the input contains object
// keys which do not match any non-ignored, exported fields in the destination.
var EnableDecoderDisallowUnknownFields = false

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return errors.New("invalid request")
	}
	if err := decodeJSON(req.Body, obj); err != nil {
		return err
	}
	values := req.URL.Query()
	if len(values.Encode()) > 0 {
		if err := mapForm(obj, values); err != nil {
			return err
		}
	}
	return nil
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	return decodeJSON(bytes.NewReader(body), obj)
}

func decodeJSON(r io.Reader, obj any) error {
	decoder := json.NewDecoder(r)
	if EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	// 附加默认值
	if err := setDefaultValue(obj); err != nil {
		return err
	}
	return validate(obj)
}

const (
	defaultTagName = "default"
)

func setDefaultValue(x interface{}) error {
	rt := reflect.TypeOf(x)
	rv := reflect.ValueOf(x)

	if rt.Kind() == reflect.Ptr {
		rt, rv = rt.Elem(), rv.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < rt.NumField(); i++ {
		rtf, rvf := rt.Field(i), rv.Field(i)
		if rtf.Anonymous && rtf.Type.Kind() == reflect.Struct {
			if err := setDefaultValue(rvf.Addr().Interface()); err != nil {
				return err
			}
		}
		if v, ok := rtf.Tag.Lookup(defaultTagName); ok && rvf.CanSet() && rvf.IsZero() {
			switch rtf.Type.Kind() {
			case reflect.Int32:
				result, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return err
				}
				rvf.SetInt(result)
			case reflect.String:
				rvf.SetString(v)
			default:

			}
		}
	}
	return nil
}
