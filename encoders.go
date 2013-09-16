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
	case *RawData:
		(v.(*RawData)).Data = &data
	default:
		return errors.New("RawDataDecoder only admits *RawData")
	}
	return nil
}

//Builds a RawDataDecoder
func NewRawDataDecoder(r io.Reader) restclient.Decoder {
	return RawDataDecoder{r}
}
