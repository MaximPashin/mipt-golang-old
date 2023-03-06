package httpfetch

import (
	"bytes"
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

func FetchAll(c *http.Client, requests []Request) []Result {
	results := make([]Result, len(requests))
	var group sync.WaitGroup
	group.Add(len(requests))
	for i := range requests {
		go fetchReq(c, &requests[i], &results[i], &group)
	}
	group.Wait()
	return results
}

func fetchReq(c *http.Client, request *Request, result *Result, group *sync.WaitGroup) {
	defer group.Done()
	req, err := http.NewRequest(request.Method, request.URL, bytes.NewReader(request.Body))
	if err != nil {
		result.Error = err
		return
	}
	fetchResult := fetch(c, req)
	result.StatusCode = fetchResult.StatusCode
	result.Error = fetchResult.Error
}

func fetch(c *http.Client, req *http.Request) *Result {
	result := &Result{}
	response, err := c.Do(req)
	if err != nil {
		result.Error = err
	} else {
		result.StatusCode = response.StatusCode
		if result.StatusCode >= 300 && result.StatusCode < 400 {
			redirect := response.Header.Get("Location")
			newReq, err := http.NewRequest(req.Method, redirect, req.Body)
			if err != nil {
				result.Error = err
			} else {
				return fetch(c, newReq)
			}
		} else {
			defer response.Body.Close()
			if _, err := ioutil.ReadAll(response.Body); err != nil {
				result.Error = err
			}
		}
	}
	return result
}
