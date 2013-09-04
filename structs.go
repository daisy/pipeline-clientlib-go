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
	Nicename string   `xml:"http://www.daisy.org/ns/pipeline/data nicename,ignoreempty"`

	Description string `xml:"http://www.daisy.org/ns/pipeline/data description,ignoreempty"`

	Homepage string   `xml:"http://www.daisy.org/ns/pipeline/data homepage,ignoreempty"`
	Inputs   []Input  `xml:"http://www.daisy.org/ns/pipeline/data input,ignoreempty"`
	Options  []Option `xml:"http://www.daisy.org/ns/pipeline/data option,ignoreempty"`
	Href     string   `xml:"href,attr"`
	Id       string   `xml:"id,attr,ignoreempty"`
}

type Option struct {
	XMLName    xml.Name `xml:"http://www.daisy.org/ns/pipeline/data option"`
	Required   bool     `xml:"required,attr,ignoreempty"`
	Sequence   bool     `xml:"sequence,attr,ignoreempty"`
	Name       string   `xml:"name,attr,ignoreempty"`
	Ordered    bool     `xml:"ordered,attr,ignoreempty"`
	Mediatype  string   `xml:"mediaType,attr,ignoreempty"`
	Desc       string   `xml:"desc,attr,ignoreempty"`
	Type       string   `xml:"type,attr,ignoreempty"`
	OutputType string   `xml:"optionType,attr,ignoreempty"`
        Separator  string   `xml:"separator,attr,ignoreempty"`
	Value      string   `xml:",chardata,ignoreempty"`
	Items      []Item

}
type Input struct {
	XMLName   xml.Name `xml:"http://www.daisy.org/ns/pipeline/data input"`
	Desc      string   `xml:"desc,attr,ignoreempty"`
	Mediatype string   `xml:"mediaType,attr,ignoreempty"`
	Name      string   `xml:"name,attr"`
	Sequence  bool     `xml:"sequence,attr,ignoreempty"`
	Items      []Item
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
	XMLName  xml.Name `xml:"http://www.daisy.org/ns/pipeline/data jobRequest"`
	Script Script `xml:"http://www.daisy.org/ns/pipeline/data script,ignoreempty"`
	Inputs []Input    `xml:"http://www.daisy.org/ns/pipeline/data input,ignoreempty"`
	Options   []Option   `xml:"http://www.daisy.org/ns/pipeline/data option,ignoreempty"`
	Callback []Callback `xml:"http://www.daisy.org/ns/pipeline/data callback,ignoreempty"`
}
