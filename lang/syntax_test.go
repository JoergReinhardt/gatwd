package lang

import (
	"fmt"
	"testing"
)

func TestAllSyntax(t *testing.T) {
	fmt.Println(AllTokens())
}
func TestParseToken(t *testing.T) {
	var str = []string{
		"=>",
		"",
		"\\x",
	}
	fmt.Println(ParseToken(str))

}
func TestMaphyntax(t *testing.T) {
	fmt.Println(AllSyntax())
}
