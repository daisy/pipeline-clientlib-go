package pipeline

import (
	"bytes"
	"testing"
)

func TestRawDataDecoder(t *testing.T) {
	msg := "heyhey"
	buf := bytes.NewBufferString(msg)
	decoder := NewRawDataDecoder(buf)
	st := RawData{Data: new([]byte)}
	err := decoder.Decode(&st)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	res := string(*(st.Data))
	if msg != res {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "msg", msg, res)
	}

}
func TestRawDataDecoderWrongType(t *testing.T) {
	msg := "heyhey"
	buf := bytes.NewBufferString(msg)
	decoder := NewRawDataDecoder(buf)
	st := RawData{Data: new([]byte)}
	err := decoder.Decode(st)
	if err == nil {
		t.Error("Expected error not thrown")
	}

}
