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
	return decodeJSON(req.Body, obj)
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
	// 附加默认值
	if err := setDefaultValue(obj); err != nil {
		return err
	}
	if err := decoder.Decode(obj); err != nil {
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
	for i := 0; i < rt.Elem().NumField(); i++ {
		rtf := rt.Elem().Field(i)
		rvf := rv.Elem().Field(i)
		if rtf.Anonymous && rtf.Type.Kind() == reflect.Struct {
			if err := setDefaultValue(rvf.Addr().Interface()); err != nil {
				return err
			}
		}
		if v, ok := rtf.Tag.Lookup(defaultTagName); ok {
			switch rtf.Type.Kind() {
			case reflect.Int32:
				result, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return err
				}
				rv.Elem().Field(i).SetInt(result)
			case reflect.String:
				rv.Elem().Field(i).SetString(v)
			default:

			}
		}
	}
	return nil
}
