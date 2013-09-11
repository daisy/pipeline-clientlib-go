package pipeline

import (
	"testing"
	"fmt"
	"bytes"
	"encoding/xml"
	"github.com/capitancambio/restclient"

)

const (
	aliveXml   = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><alive authentication='false' mode='local' version='1.6' xmlns='http://www.daisy.org/ns/pipeline/data'/>"
	scriptsXml = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><scripts href='http://localhost:8181/ws/scripts' xmlns='http://www.daisy.org/ns/pipeline/data' ><script href='http://localhost:8181/ws/scripts/zedai-to-epub3' id='zedai-to-epub3'><nicename>ZedAI to EPUB3</nicename><description>Transforms a ZedAI (DAISY 4 XML) document into an EPUB 3 publication.</description></script><script href='http://localhost:8181/ws/scripts/dtbook-to-html' id='dtbook-to-html'><nicename>DTBook to HTML</nicename><description>Transforms DTBook XML into HTML.</description></script><script href='http://localhost:8181/ws/scripts/dtbook-to-zedai' id='dtbook-to-zedai'><nicename>DTBook to ZedAI</nicename><description>Transforms DTBook XML into ZedAI XML.</description></script></scripts>"

	scriptXml = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><script href='http://localhost:8181/ws/scripts/dtbook-to-zedai' id='dtbook-to-zedai' xmlns='http://www.daisy.org/ns/pipeline/data'><nicename>DTBook to ZedAI</nicename><description>Transforms DTBook XML into ZedAI XML.</description><homepage>http://code.google.com/p/daisy-pipeline/wiki/DTBookToZedAI</homepage><input desc='One or more DTBook files to be transformed. In the case of multiple files, a merge will be performed.' mediaType='application/x-dtbook+xml' name='source' sequence='true'/><option desc='The directory to store the generated files in.' name='output-dir' ordered='true' outputType='result' required='true' sequence='false' type='anyDirURI'/></script>"
)

var expected = map[string]interface{}{
	API_ALIVE:   Alive{Mode: "local", Version: "1.6", Authentication: false},
	API_SCRIPTS: Scripts{Href: "http://localhost:8181/ws/scripts", Scripts: []Script{Script{}, Script{}, Script{}}},
	API_SCRIPT: Script{
		Href:        "http://localhost:8181/ws/scripts/dtbook-to-zedai",
		Nicename:    "DTBook to ZedAI",
		Description: "Transforms DTBook XML into ZedAI XML.",
		Homepage:    "http://code.google.com/p/daisy-pipeline/wiki/DTBookToZedAI",
		Inputs: []Input{
			Input{
				Desc:      "One or more DTBook files to be transformed. In the case of multiple files, a merge will be performed.",
				Mediatype: "application/x-dtbook+xml",
				Name:      "source",
				Sequence:  true,
			},
		},
		Options: []Option{
			Option{
				Desc:       "The directory to store the generated files in.",
				Mediatype:  "application/x-dtbook+xml",
				Name:       "output-dir",
				Required:   true,
				Sequence:   false,
				Ordered:    true,
				OutputType: "result",
				Type:       "AnyFileURI",
			},
		},
	},
        API_JOBREQUEST: JobRequest{
        },
}


type MockClient struct {
	status   int
	response string
}

func (m MockClient) Do(rr *restclient.RequestResponse) (status int, err error) {
        if m.response!=""{
                err = xml.NewDecoder(bytes.NewBufferString(m.response)).Decode(rr.Result)
        }
	return m.status, err
}
func clientMock(response string, status int) func() doer {

	return func() doer {
		return MockClient{status: status, response: response}
	}

}

func emptyClientMock() doer {
	return MockClient{status: 200, response: ""}
}
func createPipeline(maker func() doer) Pipeline {
	return Pipeline{BaseUrl: "", clientMaker: maker}
}


//Actual tests
func TestDefaultErrorHandler(t *testing.T){
	var alive Alive
	r := Pipeline{}.newResquest(API_ALIVE, &alive,nil)
        err:=defaultErrorHandler()(404,*r)

        if err.Error()!=fmt.Sprintf(ERR_404,apiEntries[API_ALIVE].urlPath){
                t.Error("Default 404 not handled")
        }

        err=defaultErrorHandler()(401,*r)
        if err.Error()!=ERR_401{
                t.Error("Default 401 not handled")
        }

        err=defaultErrorHandler()(500,*r)
        if err.Error()!=fmt.Sprintf(ERR_500," from "+apiEntries[API_ALIVE].urlPath){
                t.Error("Default 500 not handled")
        }

        r.Error.(*Error).Description="error"
        err=defaultErrorHandler()(500,*r)
        if err.Error()!=fmt.Sprintf(ERR_500,"error"){
                t.Error("Default 500 with desc not handled")
        }
        err=defaultErrorHandler()(501,*r)
        if err.Error()!=fmt.Sprintf(ERR_DEFAULT,501){
                t.Error("Default 500 with desc not handled")
        }
}

func TestCustomErrorHandler(t *testing.T){
	var alive Alive
	r := Pipeline{}.newResquest(API_ALIVE, &alive,nil)
        handler:=errorHandler(map[int]string{404: "couldnt find it"})
        err:=handler(404,*r)
        if err.Error()!="couldnt find it"{
                t.Error("custom 404 not handled")
        }
}

func TestNewRequestUnknownEntry(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Not panicked with unknown api entry")
		}
	}()
	var alive Alive
	Pipeline{}.newResquest("unknown", &alive,nil)

}

func TestNewRequestBaseUrl(t *testing.T) {
	var alive Alive
	r := Pipeline{BaseUrl: "google.com/"}.newResquest(API_ALIVE, &alive,nil)
	if r.Url != "google.com/alive" {
		t.Error("basePath not set")
	}

}

func TestNewRequestPostData(t *testing.T) {
	var alive Alive
	r := Pipeline{BaseUrl: "google.com/"}.newResquest(API_ALIVE, &alive,"data")
	if r.Data!= "data" {
		t.Error("post data not set")
	}

}
//Alive
func TestDoReq(t *testing.T) {
	var alive Alive
	pipeline := createPipeline(emptyClientMock)
	r := pipeline.newResquest(API_ALIVE, &alive,nil)
	if r.Url != "alive" {
		t.Errorf("Alive path set to %v", r.Url)
	}

}

func TestAlive(t *testing.T) {
	pipeline := createPipeline(clientMock(aliveXml, 200))
	alive, err := pipeline.Alive()
	if err != nil {
		t.Errorf("Error not nil %v", err)
	}
	exp := expected[API_ALIVE].(Alive).String()
	res := alive.String()
	if exp != res {
		t.Errorf("Alive decoding failed \nexpected %v \nresult %v", exp, res)
	}
}
func TestReqScripts(t *testing.T) {
	var scripts Scripts
	pipeline := createPipeline(emptyClientMock)
	r := pipeline.newResquest(API_SCRIPTS, &scripts,nil)
	if r.Url != "scripts" {
		t.Errorf("Scripts path set to %v", r.Url)
	}

}

func TestScripts(t *testing.T) {
	pipeline := createPipeline(clientMock(scriptsXml, 200))
	res, err := pipeline.Scripts()
	if err != nil {
		t.Errorf("Error not nil %v", err)
	}
	exp := expected[API_SCRIPTS].(Scripts)
	if exp.Href != res.Href {
		t.Errorf("Scripts decoding failed (Href)\nexpected %v \nresult %v", exp.Href, res.Href)
	}
	if len(exp.Scripts) != len(res.Scripts) {
		t.Errorf("Scripts decoding failed (Scripts len)\nexpected %v \nresult %v", len(exp.Scripts), len(res.Scripts))
	}
}

func TestReqScript(t *testing.T) {
	var script Script
	pipeline := createPipeline(emptyClientMock)
	r := pipeline.newResquest(API_SCRIPT, &script,nil, "test")
	if r.Url != "scripts/test" {
		t.Errorf("Scripts path set to %v", r.Url)
	}

}

func TestScript(t *testing.T) {
        pipeline := createPipeline(clientMock(scriptXml, 200))
        res, err := pipeline.Script("test")
        if err != nil {
        t.Errorf("Error not nil %v", err)
        }
        exp := expected[API_SCRIPT].(Script)
	if exp.Href != res.Href {
		t.Errorf("Scripts decoding failed (Href)\nexpected %v \nresult %v", exp.Href, res.Href)
	}
	if exp.Description != res.Description {
		t.Errorf("Script decoding failed (Description)\nexpected %v \nresult %v", exp.Description, res.Description)
	}
	if exp.Homepage != res.Homepage {
		t.Errorf("Scripts decoding failed (Homepage)\nexpected %v \nresult %v", exp.Homepage, res.Homepage)
	}
	if len(exp.Inputs)!= len(res.Inputs){
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Inputs), len(res.Inputs))
	}
	if len(exp.Options)!= len(res.Options){
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Options), len(res.Options))
	}
}
func TestJobReq(t *testing.T) {
        pipeline := createPipeline(clientMock("", 201))
        res, err := pipeline.JobRequest(expected[API_JOBREQUEST].(JobRequest))
        if err!=nil{
                t.Errorf("Error not nil %v", err)
        }
        if res!=true{
                t.Error("Status is not ok", err)
        }
}