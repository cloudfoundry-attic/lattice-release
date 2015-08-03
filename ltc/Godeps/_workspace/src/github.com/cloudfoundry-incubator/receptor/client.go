package receptor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/tedsuo/rata"
	"github.com/vito/go-sse/sse"
)

var ErrReadFromClosedSource = errors.New("read from closed source")
var ErrSendToClosedSource = errors.New("send to closed source")
var ErrSourceAlreadyClosed = errors.New("source already closed")
var ErrSlowConsumer = errors.New("slow consumer")

var ErrSubscribedToClosedHub = errors.New("subscribed to closed hub")
var ErrHubAlreadyClosed = errors.New("hub already closed")

const (
	ContentTypeHeader    = "Content-Type"
	XCfRouterErrorHeader = "X-Cf-Routererror"
	JSONContentType      = "application/json"
)

//go:generate counterfeiter -o fake_receptor/fake_client.go . Client

type Client interface {
	CreateTask(TaskCreateRequest) error
	Tasks() ([]TaskResponse, error)
	TasksByDomain(domain string) ([]TaskResponse, error)
	GetTask(taskId string) (TaskResponse, error)
	DeleteTask(taskId string) error
	CancelTask(taskId string) error

	CreateDesiredLRP(DesiredLRPCreateRequest) error
	GetDesiredLRP(processGuid string) (DesiredLRPResponse, error)
	UpdateDesiredLRP(processGuid string, update DesiredLRPUpdateRequest) error
	DeleteDesiredLRP(processGuid string) error
	DesiredLRPs() ([]DesiredLRPResponse, error)
	DesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error)

	ActualLRPs() ([]ActualLRPResponse, error)
	ActualLRPsByDomain(domain string) ([]ActualLRPResponse, error)
	ActualLRPsByProcessGuid(processGuid string) ([]ActualLRPResponse, error)
	ActualLRPByProcessGuidAndIndex(processGuid string, index int) (ActualLRPResponse, error)
	KillActualLRPByProcessGuidAndIndex(processGuid string, index int) error

	SubscribeToEvents() (EventSource, error)

	Cells() ([]CellResponse, error)

	UpsertDomain(domain string, ttl time.Duration) error
	Domains() ([]string, error)

	GetClient() *http.Client
	GetStreamingClient() *http.Client
}

func NewClient(url string) Client {
	return &client{
		httpClient:          cf_http.NewClient(),
		streamingHTTPClient: cf_http.NewStreamingClient(),

		reqGen: rata.NewRequestGenerator(url, Routes),
	}
}

type client struct {
	httpClient          *http.Client
	streamingHTTPClient *http.Client

	reqGen *rata.RequestGenerator
}

func (c *client) GetClient() *http.Client {
	return c.httpClient
}

func (c *client) GetStreamingClient() *http.Client {
	return c.streamingHTTPClient
}

func (c *client) CreateTask(request TaskCreateRequest) error {
	return c.doRequest(CreateTaskRoute, nil, nil, request, nil)
}

func (c *client) Tasks() ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(TasksRoute, nil, nil, nil, &tasks)
	return tasks, err
}

func (c *client) TasksByDomain(domain string) ([]TaskResponse, error) {
	tasks := []TaskResponse{}
	err := c.doRequest(TasksRoute, nil, url.Values{"domain": []string{domain}}, nil, &tasks)
	return tasks, err
}

func (c *client) GetTask(taskId string) (TaskResponse, error) {
	task := TaskResponse{}
	err := c.doRequest(GetTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, &task)
	return task, err
}

func (c *client) DeleteTask(taskId string) error {
	return c.doRequest(DeleteTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, nil)
}

func (c *client) CancelTask(taskId string) error {
	return c.doRequest(CancelTaskRoute, rata.Params{"task_guid": taskId}, nil, nil, nil)
}

func (c *client) CreateDesiredLRP(req DesiredLRPCreateRequest) error {
	return c.doRequest(CreateDesiredLRPRoute, nil, nil, req, nil)
}

func (c *client) GetDesiredLRP(processGuid string) (DesiredLRPResponse, error) {
	var desiredLRP DesiredLRPResponse
	err := c.doRequest(GetDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, nil, &desiredLRP)
	return desiredLRP, err
}

func (c *client) UpdateDesiredLRP(processGuid string, req DesiredLRPUpdateRequest) error {
	return c.doRequest(UpdateDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, req, nil)
}

func (c *client) DeleteDesiredLRP(processGuid string) error {
	return c.doRequest(DeleteDesiredLRPRoute, rata.Params{"process_guid": processGuid}, nil, nil, nil)
}

func (c *client) DesiredLRPs() ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(DesiredLRPsRoute, nil, nil, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) DesiredLRPsByDomain(domain string) ([]DesiredLRPResponse, error) {
	var desiredLRPs []DesiredLRPResponse
	err := c.doRequest(DesiredLRPsRoute, nil, url.Values{"domain": []string{domain}}, nil, &desiredLRPs)
	return desiredLRPs, err
}

func (c *client) ActualLRPs() ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsRoute, nil, nil, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPsByDomain(domain string) ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsRoute, nil, url.Values{"domain": []string{domain}}, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPsByProcessGuid(processGuid string) ([]ActualLRPResponse, error) {
	var actualLRPs []ActualLRPResponse
	err := c.doRequest(ActualLRPsByProcessGuidRoute, rata.Params{"process_guid": processGuid}, nil, nil, &actualLRPs)
	return actualLRPs, err
}

func (c *client) ActualLRPByProcessGuidAndIndex(processGuid string, index int) (ActualLRPResponse, error) {
	var actualLRP ActualLRPResponse
	err := c.doRequest(ActualLRPByProcessGuidAndIndexRoute, rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)}, nil, nil, &actualLRP)
	return actualLRP, err
}

func (c *client) KillActualLRPByProcessGuidAndIndex(processGuid string, index int) error {
	err := c.doRequest(KillActualLRPByProcessGuidAndIndexRoute, rata.Params{"process_guid": processGuid, "index": strconv.Itoa(index)}, nil, nil, nil)
	return err
}

func (c *client) SubscribeToEvents() (EventSource, error) {
	eventSource, err := sse.Connect(c.streamingHTTPClient, time.Second, func() *http.Request {
		request, err := c.reqGen.CreateRequest(EventStream, nil, nil)
		if err != nil {
			panic(err) // totally shouldn't happen
		}

		return request
	})
	if err != nil {
		return nil, err
	}

	return NewEventSource(eventSource), nil
}

func (c *client) Cells() ([]CellResponse, error) {
	var cells []CellResponse
	err := c.doRequest(CellsRoute, nil, nil, nil, &cells)
	return cells, err
}

func (c *client) UpsertDomain(domain string, ttl time.Duration) error {
	req, err := c.createRequest(UpsertDomainRoute, rata.Params{"domain": domain}, nil, nil)
	if err != nil {
		return err
	}

	if ttl != 0 {
		req.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d", int(ttl.Seconds())))
	}

	return c.do(req, nil)
}

func (c *client) Domains() ([]string, error) {
	var domains []string
	err := c.doRequest(DomainsRoute, nil, nil, nil, &domains)
	return domains, err
}

func (c *client) createRequest(requestName string, params rata.Params, queryParams url.Values, request interface{}) (*http.Request, error) {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := c.reqGen.CreateRequest(requestName, params, bytes.NewReader(requestJson))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()
	req.ContentLength = int64(len(requestJson))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *client) doRequest(requestName string, params rata.Params, queryParams url.Values, request, response interface{}) error {
	req, err := c.createRequest(requestName, params, queryParams, request)
	if err != nil {
		return err
	}
	return c.do(req, response)
}

func (c *client) do(req *http.Request, responseObject interface{}) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var parsedContentType string
	if contentType, ok := res.Header[ContentTypeHeader]; ok {
		parsedContentType, _, _ = mime.ParseMediaType(contentType[0])
	}

	if routerError, ok := res.Header[XCfRouterErrorHeader]; ok {
		return Error{Type: RouterError, Message: routerError[0]}
	}

	if parsedContentType == JSONContentType {
		return handleJSONResponse(res, responseObject)
	} else {
		return handleNonJSONResponse(res)
	}
}

func handleJSONResponse(res *http.Response, responseObject interface{}) error {
	if res.StatusCode > 299 {
		errResponse := Error{}
		if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
			return Error{Type: InvalidJSON, Message: err.Error()}
		}
		return errResponse
	}

	if err := json.NewDecoder(res.Body).Decode(responseObject); err != nil {
		return Error{Type: InvalidJSON, Message: err.Error()}
	}
	return nil
}

func handleNonJSONResponse(res *http.Response) error {
	if res.StatusCode > 299 {
		return Error{
			Type:    InvalidResponse,
			Message: fmt.Sprintf("Invalid Response with status code: %d", res.StatusCode),
		}
	}
	return nil
}
