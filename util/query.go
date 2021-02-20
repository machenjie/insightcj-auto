package util

import (
	"net/http"
	"net/url"
)

type Query struct {
	urls url.Values
}

func (q *Query) Init() *Query {
	q.urls = url.Values{}
	return q
}

func (q *Query) InitFrom(reqParams *Query) *Query {
	q.urls = url.Values{}
	q.InitFromMapList(reqParams.urls)
	return q
}

func (q *Query) InitFromMap(reqParams map[string]string) *Query {
	q.urls = url.Values{}
	for k, v := range reqParams {
		q.AddParam(k, v)
	}
	return q
}

func (q *Query) InitFromMapList(reqParams map[string][]string) *Query {
	q.urls = url.Values{}
	for k, v := range reqParams {
		for _, vin := range v {
			q.AddParam(k, vin)
		}
	}
	return q
}

func (q *Query) AddParam(property string, value string) *Query {
	if q.urls == nil {
		q.Init()
	}
	if property != "" {
		q.urls.Add(property, value)
	}
	return q
}

func (q *Query) SetToReq(req *http.Request) *Query {
	req.URL.RawQuery = q.BuildParams()
	return q
}

func (q *Query) BuildParams() string {
	if q.urls == nil {
		return ""
	}
	return q.urls.Encode()
}
