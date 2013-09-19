package pipeline

import (
	"encoding/xml"
	"errors"
	"github.com/capitancambio/restclient"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
)

type MultipartData struct {
	request JobRequest
	data    RawData
}

type MultipartEncoder struct {
	writer   io.Writer
	boundary string
}

var boundary = "pipelininginthefreeworld001"

func NewMultipartEncoder(w io.Writer) *MultipartEncoder {
	return &MultipartEncoder{w, boundary}
}
func (me MultipartEncoder) Encode(v interface{}) error {
	//get the fields
	var mpData *MultipartData
	switch v.(type) {
	case *MultipartData:
		mpData = v.(*MultipartData)
	default:
		return errors.New("MultipartEncoder only admits MultipartData")
	}

	w := multipart.NewWriter(me.writer)
	if me.boundary != "" {
		w.SetBoundary(me.boundary)
	}
	//Content-Disposition: form-data; name="job-data"; filename="/home/javi/daisy/pipeline-cli/samples/dtbook/dtbook.zip"
	//Content-Transfer-Encoding: binary
	//Content-Type: application/zip
	headerData := make(textproto.MIMEHeader)
	headerData.Add("Content-Disposition", `form-data; name="job-data"; filename="pipeline-client-go-data.zip"`)
	headerData.Add("Content-Transfer-Encoding", "binary")
	headerData.Add("Content-Type", "application/zip")

	dataWriter, err := w.CreatePart(headerData)
	if err != nil {
		return err
	}
	err = NewRawDataEncoder(dataWriter).Encode(mpData.data)
	if err != nil {
		return err
	}

	headerXml := make(textproto.MIMEHeader)
	headerXml.Add("Content-Disposition", `form-data; name="job-request"`)
	headerXml.Add("Content-Type", "application/xml; charset=utf-8")
	reqWriter, err := w.CreatePart(headerXml)
	if err != nil {
		return err
	}
	//reqWriter.Close()
	err = xml.NewEncoder(reqWriter).Encode(mpData.request)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

//Raw data struct defines a simple structure to
//store bytes
type RawData struct {
	Data *[]byte //Data
}

//Sets the bytes to the raw data structure
func (r *RawData) SetBytes(b []byte) {
	r.Data = &b
}

//Gets the bytes from the RawData structure
func (r RawData) Bytes() []byte {
	return *r.Data
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

//RawDataDecoder allows to decode raw data into a RawData structure
type RawDataEncoder struct {
	writer io.Writer
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

//Decodes the data into a raw data struct
func (d RawDataEncoder) Encode(v interface{}) error {
	var data []byte
	switch v.(type) {
	case ToBytes:
		data = (v.(ToBytes)).Bytes()
	default:
		return errors.New("RawDataDecoder only admits ToBytes interface")
	}
	_, err := d.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

//Builds a RawDataDecoder
func NewRawDataDecoder(r io.Reader) restclient.Decoder {
	return RawDataDecoder{r}
}

//Builds a RawDataDecoder
func NewRawDataEncoder(w io.Writer) restclient.Encoder {
	return RawDataEncoder{w}
}

//TODO: the following could be a hell of a library if finished
//const (
//XML  = "application/xml"
//DATA = "pplication/octet-stream"
//)

//type EncoderSupplier func(io.Writer) restclient.Encoder

////registered suppliers
//var suppliers map[string]EncoderSupplier = map[string]EncoderSupplier{
//XML:  func(w io.Writer) restclient.Encoder { return xml.NewEncoder(w) },
//DATA: NewRawDataEncoder,
//}

////type might change to sth more appropriate
////func Register(mimetype string, s EncoderSupplier) {
////suppliers[mimetype] = s
////}

////Multipart encoder encodes multypart messages
//type MultipartEncoder struct {
//w io.Writer
//}

//func (me MultipartEncoder) Encode(v interface{}) error {
////get the fields
//_, err := getFieldsMimeTypes(v)
//if err != nil {
//return err
//}
////get the encoder for each field

////write
//return nil
//}

//type multipartEntry struct {
//name     string
//mimetype string
//}

//type MPartEncoderError struct {
//Type reflect.Type
//}

//func (e MPartEncoderError) Error() string {
//return "multipart: could not encode type " + e.Type.String()
//}

//func getFieldsMimeTypes(i interface{}) (fields map[reflect.StructField]multipartEntry, err error) {
//t := reflect.TypeOf(i)
//if t == nil {
//return fields, MPartEncoderError{t}
//}
////check that t is a pointer and it's accessible
//if t.Kind() == reflect.Ptr {
////get the referenced type
//t = reflect.Indirect(reflect.ValueOf(i)).Type()
//}
//if t.Kind() != reflect.Struct {
//return fields, MPartEncoderError{t}
//}
//fields = make(map[reflect.StructField]multipartEntry)
////get the fields
//for i := 0; i < t.NumField(); i++ {
//if tag := t.Field(i).Tag; tag != "" {
//mimetype := tag.Get("mimetype")
//name := tag.Get("name")
//if name == "" && mimetype == "" {
//break
//} else if name != "" && mimetype != "" {
//fields[t.Field(i)] = multipartEntry{name, mimetype}
//}
//}
//}
//return
//}
