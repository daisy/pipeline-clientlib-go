package pipeline

import (
	"encoding/xml"
)

//Error with information from the server
type Error struct {
	XMLName     xml.Name `xml:"http://www.daisy.org/ns/pipeline/data error"`
	Description string   `xml:"http://www.daisy.org/ns/pipeline/data error>description"`
	Trace       string   `xml:"http://www.daisy.org/ns/pipeline/data error>trace"`
	Query       string   `xml:"query,attr"`
}

//More info TODO link to wiki
//Alive struct defined from the xmls
//TODO link to wiki
type Alive struct {
	XMLName        xml.Name `xml:"http://www.daisy.org/ns/pipeline/data alive"`
	Authentication bool     `xml:"authentication,attr"` //Indicates if the framework is expecting authentication
	Mode           string   `xml:"mode,attr"`           //Mode of execution
	Version        string   `xml:"version,attr"`        //Version of the pipeline framework
}

//TODO link to wiki
type Scripts struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/pipeline/data scripts"`
	Scripts []Script `xml:"http://www.daisy.org/ns/pipeline/data script"` //List of scripts available

	Href string `xml:"href,attr"` //Url used to perform this call
}

//Script struct
type Script struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data script"`
	Nicename string   `xml:"http://www.daisy.org/ns/pipeline/data nicename,omitempty"`

	Description string `xml:"http://www.daisy.org/ns/pipeline/data description,omitempty"`

	Homepage string   `xml:"http://www.daisy.org/ns/pipeline/data homepage,omitempty"`
	Inputs   []Input  `xml:"http://www.daisy.org/ns/pipeline/data input,omitempty"`
	Options  []Option `xml:"http://www.daisy.org/ns/pipeline/data option,omitempty"`
	Href     string   `xml:"href,attr"`
	Id       string   `xml:"id,attr,omitempty"`
}

type Option struct {
	XMLName    xml.Name `xml:"http://www.daisy.org/ns/pipeline/data option"`
	Required   bool     `xml:"required,attr,omitempty"`
	Sequence   bool     `xml:"sequence,attr,omitempty"`
	Name       string   `xml:"name,attr,omitempty"`
	Ordered    bool     `xml:"ordered,attr,omitempty"`
	Mediatype  string   `xml:"mediaType,attr,omitempty"`
	Desc       string   `xml:"desc,attr,omitempty"`
	Type       string   `xml:"type,attr,omitempty"`
	OutputType string   `xml:"optionType,attr,omitempty"`
	Separator  string   `xml:"separator,attr,omitempty"`
	Value      string   `xml:",chardata"`
	Items      []Item
}
type Input struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/pipeline/data input"`
	Desc      string   `xml:"desc,attr,omitempty"`
	Mediatype string   `xml:"mediaType,attr,omitempty"`
	Name      string   `xml:"name,attr"`
	Sequence  bool     `xml:"sequence,attr,omitempty"`
	Items     []Item
}

type Item struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/pipeline/data item"`
	Value   string   `xml:"value,attr"`
}

type Callback struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/pipeline/data callback"`
	Href      string   `xml:"href,attr"`
	Frequency string   `xml:"frequency,attr"`
	Type      string   `xml:"type,attr"`
}

type JobRequest struct {
	XMLName  xml.Name   `xml:"http://www.daisy.org/ns/pipeline/data jobRequest"`
	Script   Script     `xml:"http://www.daisy.org/ns/pipeline/data script,omitempty"`
	Inputs   []Input    `xml:"http://www.daisy.org/ns/pipeline/data input,omitempty"`
	Options  []Option   `xml:"http://www.daisy.org/ns/pipeline/data option,omitempty"`
	Callback []Callback `xml:"http://www.daisy.org/ns/pipeline/data callback,omitempty"`
}

type Messages struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/pipeline/data messages"`
	Message `xml:"http://www.daisy.org/ns/pipeline/data message"`
}
type Results struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data results"`
	Result   []Result `xml:"http://www.daisy.org/ns/pipeline/data result"`
	Href     string   `xml:"href,attr"`
	MimeType string   `xml:"mime-type,attr"`
}
type Job struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data job"`
	Nicename string   `xml:"http://www.daisy.org/ns/pipeline/data nicename"`
	Script   `xml:"http://www.daisy.org/ns/pipeline/data script"`
	Messages `xml:"http://www.daisy.org/ns/pipeline/data messages"`
	Log      string `xml:"http://www.daisy.org/ns/pipeline/data log"`
	Results  `xml:"http://www.daisy.org/ns/pipeline/data results"`
	Status   string `xml:"status,attr"`
	Href     string `xml:"href,attr"`
	Id       string `xml:"id,attr"`
}
type Result struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data result"`
	MimeType string   `xml:"mime-type,attr"`
	Href     string   `xml:"href,attr"`
}
type Message struct {
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data message"`
	Level    string   `xml:"level,attr"`
	Sequence string   `xml:"sequence,attr"`
}
type Log struct {
	XMLName xml.Name `xml:"http://www.daisy.org/ns/pipeline/data log"`
	Href    string   `xml:"href,attr"`
}
