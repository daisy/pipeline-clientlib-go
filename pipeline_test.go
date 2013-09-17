package pipeline

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/capitancambio/restclient"
	"io"
	"testing"
)

const (
	aliveXml   = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><alive authentication='false' mode='local' version='1.6' xmlns='http://www.daisy.org/ns/pipeline/data'/>"
	scriptsXml = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><scripts href='http://localhost:8181/ws/scripts' xmlns='http://www.daisy.org/ns/pipeline/data' ><script href='http://localhost:8181/ws/scripts/zedai-to-epub3' id='zedai-to-epub3'><nicename>ZedAI to EPUB3</nicename><description>Transforms a ZedAI (DAISY 4 XML) document into an EPUB 3 publication.</description></script><script href='http://localhost:8181/ws/scripts/dtbook-to-html' id='dtbook-to-html'><nicename>DTBook to HTML</nicename><description>Transforms DTBook XML into HTML.</description></script><script href='http://localhost:8181/ws/scripts/dtbook-to-zedai' id='dtbook-to-zedai'><nicename>DTBook to ZedAI</nicename><description>Transforms DTBook XML into ZedAI XML.</description></script></scripts>"

	scriptXml     = "<?xml version='1.0' encoding='UTF-8' standalone='no'?><script href='http://localhost:8181/ws/scripts/dtbook-to-zedai' id='dtbook-to-zedai' xmlns='http://www.daisy.org/ns/pipeline/data'><nicename>DTBook to ZedAI</nicename><description>Transforms DTBook XML into ZedAI XML.</description><homepage>http://code.google.com/p/daisy-pipeline/wiki/DTBookToZedAI</homepage><input desc='One or more DTBook files to be transformed. In the case of multiple files, a merge will be performed.' mediaType='application/x-dtbook+xml' name='source' sequence='true'/><option desc='The directory to store the generated files in.' name='output-dir' ordered='true' outputType='result' required='true' sequence='false' type='anyDirURI'/></script>"
	jobCreationOk = "<job xmlns='http://www.daisy.org/ns/pipeline/data' id='job-id-01' href='http://example.org/ws/jobs/job-id-01' status='DONE'/>"
	jobStatus     = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
                        <job xmlns="http://www.daisy.org/ns/pipeline/data" id="job-id-01" href="http://example.org/ws/jobs/job-id-01" status="DONE">
                                <!-- nicename is optional -->
                                <nicename>simple-dtbook-1</nicename>
                                <script id="dtbook-to-zedai" href="http://example.org/ws/scripts/dtbook-to-zedai">
                                        <nicename>DTBook to ZedAI</nicename>
                                        <description>Transforms DTBook XML into ZedAI XML.</description>
                                </script>
                                <messages>
                                        <message level="WARNING" sequence="22">Warning about this job</message>
                                </messages>
                                <log href="log"/>
                                <results href="http://example.org/ws/jobs/job-id-01/result" mime-type="zip">
                                        <result from="option" href="http://example.org/ws/jobs/job-id-01/result/option/output-dir" mime-type="zip" name="output-dir">
                                                <result href="http://example.org/ws/jobs/job-id-01/result/option/output-dir/file-1.xhtml" mime-type="application/xml"/>
                                        </result>
                                        <result from="port" href="http://example.org/ws/jobs/job-id-01/result/port/result" mime-type="zip" name="output-dir">
                                                <result href="http://example.org/ws/jobs/job-id-01/result/port/result/result-1.xml" mime-type="application/xml"/>
                                                <result href="http://example.org/ws/jobs/job-id-01/result/port/result/result-2.xml" mime-type="application/xml"/>
                                        </result>
                                </results>
                        </job>

                `
	errorXml = `
<?xml version="1.0" encoding="UTF-8"?>
<error query="http://localhost:8181/ws/jobs" xmlns="http://www.daisy.org/ns/pipeline/data">
    <description>Error while acquiring jobs</description>
    <trace>
    </trace>
</error>
        `
	jobsXml = `
<jobs xmlns="http://www.daisy.org/ns/pipeline/data" href="http://example.org/ws/jobs">
    <job id="job-id-01" href="http://example.org/ws/jobs/job-id-01" status="DONE">
        <nicename>job1</nicename>
    </job>
    <job id="job-id-02" href="http://example.org/ws/jobs/job-id-02" status="ERROR"/>
    <job id="job-id-03" href="http://example.org/ws/jobs/job-id-03" status="IDLE"/>
    <job id="job-id-04" href="http://example.org/ws/jobs/job-id-04" status="RUNNING">
        <nicename>job4</nicename>
    </job>
</jobs>
`
	T_STRING = "Wrong %v\nexpected: %v\nresult:%v\n"
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
	API_JOBREQUEST: JobRequest{},

	API_JOB: Job{
		Id:       "job-id-01",
		Status:   "DONE",
		Nicename: "simple-dtbook-1",
		Log:      Log{Href: "log"},
	},
}

type MockClient struct {
	status          int
	response        string
	EncoderSupplier func(io.Writer) restclient.Encoder //Supplies the endoder objects
	DecoderSupplier func(io.Reader) restclient.Decoder //Supplies the endoder objects
	fail            bool
}

func (m *MockClient) SetDecoderSupplier(fn func(io.Reader) restclient.Decoder) {
	m.DecoderSupplier = fn
}
func (m MockClient) Do(rr *restclient.RequestResponse) (status int, err error) {
	if m.response != "" {
		err = m.DecoderSupplier(bytes.NewBufferString(m.response)).Decode(rr.Result)
	}
	if m.fail {
		err = m.DecoderSupplier(bytes.NewBufferString(errorXml)).Decode(rr.Error)
		if err != nil {
			println("THIS errorXml SHOULD NOT HAPPEN")
			panic(err.Error())
		}
	}
	return m.status, err
}
func clientMock(response string, status int) func() doer {

	return func() doer {
		return &MockClient{
			status:   status,
			response: response,
			EncoderSupplier: func(w io.Writer) restclient.Encoder {
				return xml.NewEncoder(w)
			},
			DecoderSupplier: func(r io.Reader) restclient.Decoder {
				return xml.NewDecoder(r)
			},
			fail: false,
		}
	}

}

func failingMock() func() doer {

	return func() doer {
		return &MockClient{
			status:   200,
			response: "",
			EncoderSupplier: func(w io.Writer) restclient.Encoder {
				return xml.NewEncoder(w)
			},
			DecoderSupplier: func(r io.Reader) restclient.Decoder {
				return xml.NewDecoder(r)
			},
			fail: true,
		}
	}

}

func emptyClientMock() doer {
	return &MockClient{status: 200, response: ""}
}
func createPipeline(maker func() doer) Pipeline {
	return Pipeline{BaseUrl: "base/", clientMaker: maker}
}

//Actual tests
func TestDefaultErrorHandler(t *testing.T) {
	var alive Alive
	r := Pipeline{}.newResquest(API_ALIVE, &alive, nil)
	err := defaultErrorHandler()(404, *r)

	if err.Error() != fmt.Sprintf(ERR_404, apiEntries[API_ALIVE].urlPath) {
		t.Error("Default 404 not handled")
	}

	err = defaultErrorHandler()(401, *r)
	if err.Error() != ERR_401 {
		t.Error("Default 401 not handled")
	}

	err = defaultErrorHandler()(500, *r)
	if err.Error() != fmt.Sprintf(ERR_500, " from "+apiEntries[API_ALIVE].urlPath) {
		t.Error("Default 500 not handled")
	}

	r.Error.(*Error).Description = "error"
	err = defaultErrorHandler()(500, *r)
	if err.Error() != fmt.Sprintf(ERR_500, "error") {
		t.Error("Default 500 with desc not handled")
	}
	err = defaultErrorHandler()(501, *r)
	if err.Error() != fmt.Sprintf(ERR_DEFAULT, 501) {
		t.Error("Default 500 with desc not handled")
	}
}

func TestCustomErrorHandler(t *testing.T) {
	var alive Alive
	r := Pipeline{}.newResquest(API_ALIVE, &alive, nil)
	handler := errorHandler(map[int]string{404: "couldnt find it"})
	err := handler(404, *r)
	if err.Error() != "couldnt find it" {
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
	Pipeline{}.newResquest("unknown", &alive, nil)

}

func TestNewRequestBaseUrl(t *testing.T) {
	var alive Alive
	r := Pipeline{BaseUrl: "google.com/"}.newResquest(API_ALIVE, &alive, nil)
	if r.Url != "google.com/alive" {
		t.Error("basePath not set")
	}

}

func TestNewRequestPostData(t *testing.T) {
	var alive Alive
	r := Pipeline{BaseUrl: "google.com/"}.newResquest(API_ALIVE, &alive, "data")
	if r.Data != "data" {
		t.Error("post data not set")
	}

}

//Alive
func TestDoReq(t *testing.T) {
	var alive Alive
	pipeline := createPipeline(emptyClientMock)
	r := pipeline.newResquest(API_ALIVE, &alive, nil)
	if r.Url != "base/alive" {
		t.Errorf("Alive path set to %v", r.Url)
	}

}

func TestServerError(t *testing.T) {
	cli := failingMock()
	pipeline := createPipeline(cli)
	_, err := pipeline.Alive()
	if err == nil {
		t.Errorf("Exepecte error not thrown")
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
	r := pipeline.newResquest(API_SCRIPTS, &scripts, nil)
	if r.Url != "base/scripts" {
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
	r := pipeline.newResquest(API_SCRIPT, &script, nil, "test")
	if r.Url != "base/scripts/test" {
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
	if len(exp.Inputs) != len(res.Inputs) {
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Inputs), len(res.Inputs))
	}
	if len(exp.Options) != len(res.Options) {
		t.Errorf("Scripts decoding failed (inputs)\nexpected %v \nresult %v", len(exp.Options), len(res.Options))
	}
}
func TestJobReq(t *testing.T) {
	pipeline := createPipeline(clientMock(jobCreationOk, 201))
	res, err := pipeline.JobRequest(expected[API_JOBREQUEST].(JobRequest))
	if err != nil {
		t.Errorf("Error not nil %v", err)
	}
	if res.Id == "" {
		t.Error("job id not ok", err)
	}
}

func TestScriptUrl(t *testing.T) {
	pipeline := createPipeline(clientMock("", 0))
	url := pipeline.ScriptUrl("unpalo")
	if url != "base/scripts/unpalo" {
		t.Errorf("Script url \nexpected %v \nresult %v", "base/scripts/unpalo", url)
	}
}

func TestJob(t *testing.T) {
	pipeline := createPipeline(clientMock(jobStatus, 200))
	res, err := pipeline.Job("jobId", 0)
	expJob := expected[API_JOB].(Job)
	if err != nil {
		t.Errorf("Error not nil %v", err)
	}

	if expJob.Id != res.Id {
		t.Errorf(T_STRING, "id", expJob.Id, res.Id)
	}

	if expJob.Nicename != res.Nicename {
		t.Errorf(T_STRING, "nicename", expJob.Id, res.Id)
	}
	if expJob.Log.Href != res.Log.Href {
		t.Errorf(T_STRING, "log", expJob.Id, res.Id)
	}
	if len(res.Results.Result) != 2 {
		t.Errorf(T_STRING, "results len", 2, len(res.Results.Result))
	}
	if len(res.Results.Result[0].Result) != 1 {
		t.Errorf(T_STRING, "results len", 1, len(res.Results.Result[0].Result))
	}
	if len(res.Results.Result[1].Result) != 2 {
		t.Errorf(T_STRING, "results len", 2, len(res.Results.Result[1].Result))
	}
}

func TestResults(t *testing.T) {
	msg := "learn to swim"
	pipeline := createPipeline(clientMock(msg, 200))
	data, err := pipeline.Results("id")
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	res := string(data)
	if msg != res {
		t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "msg ", msg, res)
	}
}

func TestJobs(t *testing.T) {
	pipeline := createPipeline(clientMock(jobsXml, 200))
	res, err := pipeline.Jobs()
	idTemp := "job-id-0%v"
	if err != nil {
		t.Errorf("Error not nil %v", err)
	}
	if len(res.Jobs) != 4 {
		t.Errorf("Wrong jobs size", res.Jobs)
	}
	for idx, job := range res.Jobs {
		jobId := fmt.Sprintf(idTemp, idx+1)
		if jobId != job.Id {
			t.Errorf("Wrong %v\n\tExpected: %v\n\tResult: %v", "jobId ", jobId, job.Id)
		}
	}

}
