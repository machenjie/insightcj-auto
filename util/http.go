package util

import (
	"bytes"
	"encoding/xml"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"
)

type EasyRequest struct {
	Request   *http.Request
	Transport *http.Transport
	Query     *Query
	Header    *Header
	Jar       http.CookieJar
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var defaultTransport = &http.Transport{MaxIdleConnsPerHost: runtime.NumCPU(), ResponseHeaderTimeout: time.Minute}
var httpClient = &http.Client{Transport: defaultTransport}

func (er *EasyRequest) SetQuery(q *Query) *EasyRequest {
	er.Query = q
	return er
}

func (er *EasyRequest) SetHeader(h *Header) *EasyRequest {
	er.Header = h
	return er
}

func (er *EasyRequest) SetJar(jar http.CookieJar) *EasyRequest {
	er.Jar = jar
	return er
}

func (er *EasyRequest) SetRequest(request *http.Request) *EasyRequest {
	er.Request = request
	return er
}

func (er *EasyRequest) SetTransport(transport *http.Transport) *EasyRequest {
	er.Transport = transport
	return er
}

func (er *EasyRequest) prepareRequest() *http.Client {
	client := httpClient
	if er.Transport != nil || er.Jar != nil {
		t := er.Transport
		if t == nil {
			t = defaultTransport
		}
		client = &http.Client{Transport: t, Jar: er.Jar}
	}
	if er.Query != nil {
		er.Query.SetToReq(er.Request)
	}
	if er.Header != nil {
		er.Header.SetToReq(er.Request)
	}
	return client
}

func (er *EasyRequest) Stream() io.ReadCloser {
	if er.Request == nil {
		Log.Warnf("Stream Request error, no valid request")
		return nil
	}
	client := er.prepareRequest()
	resp, err := client.Do(er.Request)
	if err != nil {
		Log.Warnf("Stream Request error [[%+v]]! request info[[%+v]]", err, er.Request)
		return nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		respBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		Log.Warnf("Bad Stream request! Response code: [[%d]], response body: [[%s]]", resp.StatusCode, string(respBody))
		return nil
	}
	return resp.Body
}

func (er *EasyRequest) Do() []byte {
	if er.Request == nil {
		Log.Warnf("Request error, no valid request")
		return nil
	}
	client := er.prepareRequest()
	resp, err := client.Do(er.Request)
	if err != nil {
		Log.Warnf("Request of %s error [[%+v]]!", er.Request.URL.String(), err)
		return nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		Log.Warnf("Read response body of %s error [[%+v]]! response info [[%s]]", er.Request.URL.String(), err, string(respBody))
		return nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		Log.Warnf("Bad request of %s! Response code: [[%d]], response body: [[%v]]", er.Request.URL.String(), resp.StatusCode, string(respBody))
		return nil
	}
	return respBody
}

func (er *EasyRequest) ToJson(js interface{}) bool {
	respBody := er.Do()
	if respBody == nil {
		return false
	}
	if !jsoniter.Valid(respBody) {
		Log.Warnf("Response body [[%s]] is not json format!", string(respBody))
		return false
	}
	err := json.Unmarshal(respBody, js)
	if err != nil {
		Log.Warnf("Response body unmarshal error [[%+v]]!", err)
		return false
	}
	return true
}

func (er *EasyRequest) ToHtml(x interface{}) bool {
	respBody := er.Do()
	if respBody == nil {
		return false
	}
	xmlDecoder := xml.NewDecoder(bytes.NewReader(respBody))
	xmlDecoder.Strict = false
	xmlDecoder.AutoClose = xml.HTMLAutoClose
	xmlDecoder.Entity = xml.HTMLEntity
	err := xmlDecoder.Decode(x)
	if err != nil {
		Log.Warnf("Response body unmarshal error [[%+v]]!", err)
		return false
	}
	return true
}

func (er *EasyRequest) ToXml(x interface{}) bool {
	respBody := er.Do()
	if respBody == nil {
		return false
	}
	err := xml.Unmarshal(respBody, x)
	if err != nil {
		Log.Warnf("Response body unmarshal error [[%+v]]!", err)
		return false
	}
	return true
}
