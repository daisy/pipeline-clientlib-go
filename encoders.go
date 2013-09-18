package pipeline

import (
	"errors"
	"github.com/capitancambio/restclient"
	"io"
	"io/ioutil"
)


//Raw data struct defines a simple structure to
//store bytes
type RawData struct {
	Data *[]byte //Data
}

func (r *RawData) SetBytes(b []byte) {
	r.Data = &b
}

type ToBytes interface {
	Bytes() []byte
}

type FromBytes interface {
	SetBytes([]byte)
}

//RawDataDecoder allows to decode raw data into a RawData structure
type RawDataDecoder struct {
	reader io.Reader
}

//Decodes the data into a raw data struct
func (d RawDataDecoder) Decode(v interface{}) error {
	data, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return err
	}
	switch v.(type) {
	case FromBytes:
		(v.(FromBytes)).SetBytes(data)
	default:
		return errors.New("RawDataDecoder only admits FromBytes interface")
	}
	return nil
}

//Builds a RawDataDecoder
func NewRawDataDecoder(r io.Reader) restclient.Decoder {
	return RawDataDecoder{r}
}
