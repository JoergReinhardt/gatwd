package types

import (
	"strings"
)

var str = strings.Split("this is a public service announcement..."+
	" and this is not a test!"+
	" and a hip, hop and'a hippe'd hop...", " ")
var data = func() []Data {
	var data = []Data{}
	for _, d := range str {
		data = append(data, conData(d))
	}
	return data
}()
