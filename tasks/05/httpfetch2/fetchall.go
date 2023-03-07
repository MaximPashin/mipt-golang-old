package httpfetch2

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
)

type Request struct {
	Method string
	URL    string
	Body   []byte
}

type Result struct {
	StatusCode int
	Error      error
}

var errRedirect = errors.New("Redirect")

func FetchAll(ctx context.Context, c *http.Client, requests <-chan Request) <-chan Result {
	ch := make(chan Result)
	var wg sync.WaitGroup
	go func() {
		defer func() {
			wg.Wait()
			close(ch)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case request, more := <-requests:
				if more == true {
					go func() {
						wg.Add(1)
						result := &Result{}
						httpRequest, err := buildHTTPRequest(&ctx, &request)
						if err != nil {
							result = &Result{Error: err}
						} else {
							result = fetch(c, httpRequest)
						}
						ch <- *result
						wg.Done()
					}()
				} else {
					return
				}
			}
		}
	}()
	return ch
}

func fetch(c *http.Client, req *http.Request) *Result {
	select {
	case <-req.Context().Done():
		return &Result{Error: context.Canceled}
	default:
	}
	var result *Result
	response, err := c.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		result = &Result{Error: err}
	} else {
		result, err = buildResult(response)
		if err != nil {
			switch {
			case errors.Is(err, errRedirect):
				result = handleRedirect(c, response)
			default:
				result = &Result{Error: err}
			}
		}
	}
	return result
}

func buildHTTPRequest(ctx *context.Context, request *Request) (*http.Request, error) {
	select {
	case <-(*ctx).Done():
		return nil, context.Canceled
	default:
		return http.NewRequestWithContext(*ctx, request.Method, request.URL, bytes.NewReader(request.Body))
	}
}

func buildResult(response *http.Response) (*Result, error) {
	result := &Result{}
	result.StatusCode = response.StatusCode
	if result.StatusCode >= 300 && result.StatusCode < 400 {
		return nil, errRedirect
	} else {
		if _, err := ioutil.ReadAll(response.Body); err != nil {
			result.Error = err
		}
	}
	return result, nil
}

func handleRedirect(c *http.Client, response *http.Response) *Result {
	redirect := response.Header.Get("Location")
	newReq, err := http.NewRequest(response.Request.Method, redirect, response.Request.Body)
	if err != nil {
		return &Result{Error: err}
	} else {
		return fetch(c, newReq)
	}
}
