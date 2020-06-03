package main

import (
	"database/sql"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func sqlRowJSON(row map[string]sql.NullString) string {
	m := map[string]interface{}{}
	for k, v := range row {
		if v.Valid {
			m[k] = v.String
		} else {
			m[k] = nil
		}
	}
	return jsonEncode(m)
}

func sqlRowsJSON(rows []map[string]sql.NullString) string {
	ms := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		m := map[string]interface{}{}
		for k, v := range row {
			if v.Valid {
				m[k] = v.String
			} else {
				m[k] = nil
			}
		}
		ms[i] = m
	}
	return jsonEncode(ms)
}

func jsonEncode(v interface{}) (s string) {
	s, _ = json.MarshalToString(v)
	return
}
