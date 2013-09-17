package pipeline

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/capitancambio/restclient"
	"io"
	"log"
)

//Available api entry names
const (
	API_ALIVE      = "alive"
	API_SCRIPT     = "script"
	API_SCRIPTS    = "scripts"
	API_JOBREQUEST = "jobRequest"
	API_JOB        = "jobs"
	API_JOBS       = "jobs"
	API_DEL_JOB    = "del_job"
	API_RESULT     = "results"
)

//Error messages
const (
	ERR_404     = "Resource not found %v"
	ERR_401     = "You don't have enough permissions, check your configuration"
	ERR_500     = "Server error: %v"
	ERR_DEFAULT = "Framework server error (code: %v)"
)

//Defines the information for an api entry
type apiEntry struct {
	urlPath  string
	method   string
	okStatus int
}

//Available api entries
var apiEntries = map[string]apiEntry{
	API_ALIVE:      apiEntry{"alive", "GET", 200},
	API_SCRIPTS:    apiEntry{"scripts", "GET", 200},
	API_SCRIPT:     apiEntry{"scripts/%v", "GET", 200},
	API_JOBREQUEST: apiEntry{"jobs", "POST", 201},
	API_JOB:        apiEntry{"jobs/%v?msgSeq=%v", "GET", 200},
	API_DEL_JOB:    apiEntry{"jobs/%v", "DELETE", 204},
	API_RESULT:     apiEntry{"jobs/%v/result", "GET", 200},
}

//Default error handler has generic treatment for errors derived from the http status
func defaultErrorHandler() func(status int, respose restclient.RequestResponse) error {
	return errorHandler(make(map[int]string))
}

//Returns an error handler adding specific treatments to different status apart from the ones defined in the default
func errorHandler(handlers map[int]string) func(status int, respose restclient.RequestResponse) error {
	return func(status int, req restclient.RequestResponse) error {
		if err, ok := handlers[status]; ok {
			return errors.New(err)
		}
		switch status {
		case 404:
			return fmt.Errorf(ERR_404, req.Url)
		case 401:
			return errors.New(ERR_401)
		case 500: //check response from the server
			if req.Error.(*Error).Description != "" {
				return fmt.Errorf(ERR_500, req.Error.(*Error).Description)
			} else {
				return fmt.Errorf(ERR_500, " from "+req.Url)
			}
		}
		return fmt.Errorf(ERR_DEFAULT, status)
	}
}

//Pipeline struct stores different configuration paramenters
//for the communication with the pipeline framework
type Pipeline struct {
	BaseUrl     string      //baseurl of the framework
	clientMaker func() doer //client to perform the rest queries
}

func NewPipeline(baseUrl string) *Pipeline {
	return &Pipeline{
		BaseUrl:     baseUrl,
		clientMaker: newClient,
	}
}

//Convinience interface for testing
type doer interface {
	Do(*restclient.RequestResponse) (status int, err error)
	SetDecoderSupplier(func(io.Reader) restclient.Decoder)
}

//Creates a new client setting the correct encoders
func newClient() doer {
	client := restclient.New()
	client.EncoderSupplier = func(w io.Writer) restclient.Encoder {
		return xml.NewEncoder(w)
	}
	client.DecoderSupplier = func(r io.Reader) restclient.Decoder {
		return xml.NewDecoder(r)
	}
	return client
}

//Creates a new request object for the api entry and the target struct where the response for the sever will be decoded
func (p Pipeline) newResquest(apiEntry string, targetPtr interface{}, postData interface{}, args ...interface{}) *restclient.RequestResponse {

	if entry, ok := apiEntries[apiEntry]; ok {
		url := p.BaseUrl + entry.urlPath
		if len(args) > 0 {
			url = fmt.Sprintf(url, args...)
		}
		r := &restclient.RequestResponse{
			Url:            url,
			Method:         entry.method,
			Result:         targetPtr,
			Error:          &Error{},
			ExpectedStatus: entry.okStatus,
			Data:           postData,
		}

		return r
	} else {
		panic(fmt.Sprintf("No api entry found for %v ", apiEntry))
	}
}

//Executes the request against the client
func (p Pipeline) do(req *restclient.RequestResponse, handler func(int, restclient.RequestResponse) error) (status int, err error) {
	status, err = p.clientMaker().Do(req)
	if err != nil {
		if err == restclient.UnexpectedStatus {
			err = handler(status, *req)
		}
		return
	}
	errStr := req.Error.(*Error).Description
	if errStr != "" {
		return status, fmt.Errorf("WS ERROR: %v", errStr)
	}
	return
}

//Returns a simple string representation of the Alive struct in the format:
//Alive:[#authentication:value #mode:value #version:value]
func (a Alive) String() string {
	return fmt.Sprintf("Alive:[#authentication:%v #mode:%v #version:%v]", a.Authentication, a.Mode, a.Version)
}

//Calls the alive api entry
//TODO link to wiki
func (p Pipeline) Alive() (alive Alive, err error) {
	req := p.newResquest(API_ALIVE, &alive, nil)
	_, err = p.do(req, defaultErrorHandler())
	return
}

//List of scripts

//Returns the list of available scripts
func (p Pipeline) Scripts() (scripts Scripts, err error) {
	req := p.newResquest(API_SCRIPTS, &scripts, nil)
	_, err = p.do(req, defaultErrorHandler())
	return
}

//Returns a simple string representation of the Scripts struct in the format:
//Scripts:[#href:value #size:value]
func (s Scripts) String() string {
	return fmt.Sprintf("Scripts:[#href:%v #size:%v]", s.Href, len(s.Scripts))
}

//Returns the list of available scripts
func (p Pipeline) Script(id string) (script Script, err error) {
	req := p.newResquest(API_SCRIPT, &script, nil, id)
	_, err = p.do(req, errorHandler(map[int]string{404: "Script " + id + " not found"}))
	return
}
func (p Pipeline) ScriptUrl(id string) string {
	req := p.newResquest(API_SCRIPT, nil, nil, id)
	return req.Url
}

//JobRequest

func (p Pipeline) JobRequest(newJob JobRequest) (job Job, err error) {
	log.Println("Sending job request")
	log.Println(newJob.Script.Id)
	req := p.newResquest(API_JOBREQUEST, &job, &newJob)
	_, err = p.do(req, errorHandler(map[int]string{
		400: "Job request is not valid",
	}))
	return
}

func (p Pipeline) Job(id string, messageSequence int) (job Job, err error) {
	req := p.newResquest(API_JOB, &job, nil, id, messageSequence)
	_, err = p.do(req, errorHandler(map[int]string{
		404: "Job " + id + " not found",
	}))
	return
}

func (p Pipeline) Jobs() (jobs Jobs, err error) {
	req := p.newResquest(API_JOBS, &jobs, nil)
	_, err = p.do(req, defaultErrorHandler())
	return
}

func (p Pipeline) DeleteJob(id string) (ok bool, err error) {
	req := p.newResquest(API_DEL_JOB, nil, nil, id)
	_, err = p.do(req, errorHandler(map[int]string{
		404: "Job " + id + " not found",
	}))
	if err == nil {
		ok = true
	}
	return
}

//Overrides the xml decoder to get raw data
func resultClientMaker(p Pipeline) func() doer {
	return func() doer {
		cli := p.clientMaker()
		cli.SetDecoderSupplier(func(r io.Reader) restclient.Decoder {
			return NewRawDataDecoder(r)
		})
		return cli
	}
}

//return the results as an array of bytes
func (p Pipeline) Results(id string) (data []byte, err error) {
	//override the client maker
	p.clientMaker = resultClientMaker(p)
	rd := &RawData{Data: new([]byte)}
	req := p.newResquest(API_RESULT, rd, nil, id)
	_, err = p.do(req, errorHandler(map[int]string{
		404: "Job " + id + " not found",
	}))
	if err != nil {
		return nil, err
	}
	return *(rd.Data), nil
}
