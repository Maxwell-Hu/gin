// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"bytes"
	"encoding/xml"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"

	"gopkg.in/bluesuncorp/validator.v5"
)

func WrapF(f http.HandlerFunc) HandlerFunc {
	return func(c *Context) {
		f(c.Writer, c.Request)
	}
}

func WrapH(h http.Handler) HandlerFunc {
	return func(c *Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

type H map[string]interface{}

// Allows type H to be used with xml.Marshal
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	if err := e.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return err
	}
	return nil
}

func parseBindError(err error) error {
	switch err.(type) {
	case *validator.StructErrors:
		unwrapped := err.(*validator.StructErrors)
		fields := listOfField(unwrapped.Errors)
		humanError := tohuman(fields)
		return &Error{
			Err:  unwrapped,
			Type: ErrorTypeBind,
			Meta: H{
				"message": humanError,
				"fields":  fields,
			},
		}
	default:
		return err
	}
}

func listOfField(errors map[string]*validator.FieldError) []string {
	fields := make([]string, 0, len(errors))
	for key := range errors {
		fields = append(fields, strings.ToLower(key))
	}
	return fields
}

func tohuman(fields []string) string {
	length := len(fields)
	var buf bytes.Buffer
	if length > 1 {
		buf.WriteString(strings.Join(fields[:length-1], ", "))
		buf.WriteString(" and ")
	}
	buf.WriteString(fields[length-1])
	if len(fields) == 1 {
		buf.WriteString(" is ")
	} else {
		buf.WriteString(" are ")
	}
	buf.WriteString("required.")
	return buf.String()
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

func chooseData(custom, wildcard interface{}) interface{} {
	if custom == nil {
		if wildcard == nil {
			panic("negotiation config is invalid")
		}
		return wildcard
	}
	return custom
}

func parseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		index := strings.IndexByte(part, ';')
		if index >= 0 {
			part = part[0:index]
		}
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			out = append(out, part)
		}
	}
	return out
}

func lastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	if len(relativePath) == 0 {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}
