package utils

import (
	"gdbc/constant"
	"testing"
	"time"
)

func Test_LenencIntDecode(t *testing.T) {

}

func Test_LenencIntEncode(t *testing.T) {

}

func Test_LenencStringDecode(t *testing.T) {

}

//go test -v -run="Test_DriverValueHandler"
func Test_DriverValueHandler(t *testing.T) {
	t.Log(DriverValueHandler("arg", constant.TextProtocol))
	t.Log(DriverValueHandler(int64(2<<32), constant.TextProtocol))
	t.Log(DriverValueHandler(float64(1.3149526), constant.TextProtocol))
	t.Log(DriverValueHandler([]byte{'a', 0xff}, constant.TextProtocol))
	t.Log(DriverValueHandler(time.Now(), constant.TextProtocol))
}

func Test_FormatBinaryDateTime(t *testing.T) {
	data := []byte{0x0b, 0xda, 0x07, 0x0a, 0x11, 0x13, 0x1b, 0x1e, 0x01, 0x00, 0x00, 0x00}
	n := 11
	dateTime, e := FormatBinaryDateTime(n, data)
	if e != nil {
		t.Error(e)
		t.Fail()
	}
	t.Log(dateTime.String())
}
