package watoken

import (
	"os"
	"testing"
)

func TestEncode(t *testing.T) {
	privkey := os.Getenv("PRIVATEKEY")
	str, _ := EncodeforHours("62895601060000", "Helpdesk Pamong Desa", privkey, 43830)
	println(privkey)
	println(str)
	atr, _ := DecodeGetId("", str)
	println(atr)

}
