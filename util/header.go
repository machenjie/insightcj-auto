package util

import (
	"net/http"
)

type Header struct {
	header http.Header
}

func (h *Header) Init() *Header {
	h.header = map[string][]string{}
	return h
}

func (h *Header) InitFrom(header *Header) *Header {
	h.header = map[string][]string{}
	h.InitFromMapList(header.header)
	return h
}

func (h *Header) InitFromMap(header map[string]string) *Header {
	h.header = http.Header{}
	for k, v := range header {
		h.header.Set(k, v)
	}
	return h
}

func (h *Header) InitFromMapList(header map[string][]string) *Header {
	h.header = http.Header{}
	for k, v := range header {
		for _, vin := range v {
			h.AddHeader(k, vin)
		}
	}
	return h
}

func (h *Header) AddHeader(field string, value string) *Header {
	if h.header == nil {
		h.Init()
	}
	if field != "" {
		h.header.Add(field, value)
	}
	return h
}

func (h *Header) SetToReq(req *http.Request) *Header {
	req.Header = h.BuildHeaders()
	return h
}

func (h *Header) BuildHeaders() http.Header {
	return h.header.Clone()
}
