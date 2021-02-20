package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/machenjie/insightcj-auto/util"
	"golang.org/x/net/publicsuffix"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
	"strings"
	"time"
)

var maxResponseCount = 200
var loginName = ""
var loginPassword = ""
var txtResponse = `我们在信标链上添加了随机取样的 “同步委员会”。这样做的目的是让轻客户端以较低的开销 (每天至少需要约20KB来保持，需要约500个字节来确定单个区块) 来确定信标链头。这将使得轻客户端实际上可用于移动设备、信标链 之类的浏览器内的应用案例 (以及合并后的整个以太坊)，从而为更加去信任的钱包生态打好基础。
在每个时间段（约27小时）内，随机选择 1024 位验证者作为同步委员会的成员。同步委员会中的验证者将发布证明当前链头的签名。这些签名将作为LightClientUpdate对象的一部分被广播至区块链，这可以帮助轻客户端找到链头；并且签名会被打包进链，验证者会分得奖励。
主要PR：
https://github.com/ethereum/eth2.0-specs/pull/2130
核算改革 (第一层)
给验证者的奖励不再通过计算得出。此前，我们的方法为存储PendingAttestation对象然后在最后对它们进行处理。而现在我们添加了一个位字段以存储每个验证者的状态，从而可以实时收集参与数据。位字段按照“混洗”的方法进行排序，以确保同一个委员会的验证者的记录同时显示。这一改变的目的是简化客户端实现，并使得更新默克尔树的成本更低。
主要PR：
https://github.com/ethereum/eth2.0-specs/pull/2176
核算改革 (第二层)
我们每 64 个 epochs 更新一次验证者集并进行一次惩罚核算，而不再每个 epoch都计算一次。这样做是为了极大地降低处理“空时段过渡 (empty epoch transitions)”的复杂性——比如，在一条参与率非常低的链中，两个相继的区块之间隔了一千个 slot，其间仅有空块。目前为了处理这样的链，客户端们将需要每个epoch重新计算一次验证者的余额以对验证者执行怠工惩罚。而这项提案应用之后，客户端仅需要每隔 64 个 epoch 核算一次。
此外，我们对怠工惩罚 (inactivity leaks) 增加了两项变动：
1. 每个验证者的怠工惩罚力度降低至1/4。也就是说，如果链上出现怠工惩罚，当一个完全离线的验证者损失其余额的~10%的数额时，在此期间另一个90%都在线的验证者仅损失其余额的~0.1% (而不是~1%)。这样做是为了加大对作恶节点的惩罚力度，对那些仅仅由于网络连接不佳而掉线的验证者则降低惩罚力度。点进链接查看更多的讨论
2. 区块敲定后怠工惩罚会逐渐减少，而不会停止。即区块被敲定后，离线节点的余额将持续减少，这样确保了参与率显著高于2/3，而不是刚刚超过阈值。点进链接查看更多的讨论 (不过请注意与此处略有不同)。
主要PRs：
https://github.com/ethereum/eth2.0-specs/pull/2192
https://github.com/ethereum/eth2.0-specs/pull/2194
惩罚常数调整
很庆幸，尽管我们还没有完全解决验证者惩罚的问题，但在某种程度上已经摆脱了困境。我们会改变以下常数：
目前，如果在最近的 slot 里没有区块发布，那么出于 LMD GHOST 证明的目的，该 slot 里面的证明会被算作支持证明者所支持的最近区块。例如，在下图，空白 (BLANK) 区块的证明也会算入 A 的证明里。
但是，这容易招致 34% 攻击。如果有m名验证者被分配到每个 slot，那么一个恶意攻击者就可以控制每个 slot 的0.34 * m。攻击是这样进行的：攻击者不发布 B，且不发布任何他们的证明。所有的诚实证明者对他们在slotn看到A、在slotn+1什么都没看到的声明进行投票，在slot n+2，诚实提议者会在区块A上生成区块C，而诚实的验证者们会支持C。此时，恶意提议者发布B并对slot n+1和n+2做证明。这样，底部分叉有0.68 * m的验证者支持它，而顶部分叉只有0.66 * m的验证者支持，由此底部分叉胜出。`

func parseCSRF(html string) (name, token string, ok bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		util.Log.Errorf("Parse csrf from [[%s]] failed, error %+v!", html, err)
		return
	}
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		metaName, nameExist := s.Attr("name")
		if nameExist {
			if metaName == "csrf-token" {
				token, _ = s.Attr("content")
			} else if metaName == "csrf-param" {
				name, _ = s.Attr("content")
			}
		}
	})
	if token != "" && name != "" {
		ok = true
	}
	return
}

func easyRequest() *util.EasyRequest {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		jar = nil
	}
	er := &util.EasyRequest{
		Transport: &http.Transport{MaxIdleConnsPerHost: runtime.NumCPU(), ResponseHeaderTimeout: 30 * time.Second},
		Jar:       jar,
	}
	return er
}

func login() *util.EasyRequest {
	loginUrl := "https://insightcj.com/signin"
	loginGetReq, err := http.NewRequest("GET", loginUrl, nil)
	if err != nil {
		util.Log.Errorf("generate insightcj login request failed, error %+v!", err)
		return nil
	}
	er := easyRequest()
	er.SetHeader((&util.Header{}).InitFromMap(map[string]string{
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Host":       "insightcj.com",
		"Origin":     "https://insightcj.com",
		"Referer":    "https://insightcj.com/signin",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 Edg/85.0.564.63",
	}))
	resp := er.SetRequest(loginGetReq).Do()
	if resp == nil {
		util.Log.Errorf("do insightcj login request failed!")
		return nil
	}
	csrfName, csrfToken, ok := parseCSRF(string(resp))
	if !ok {
		util.Log.Errorf("parse csrf from insightcj login html failed!")
		return nil
	}
	loginBody := fmt.Sprintf("name=%s&pass=%s&%s=%s", url.QueryEscape(loginName), url.QueryEscape(loginPassword), url.QueryEscape(csrfName), url.QueryEscape(csrfToken))
	loginPostReq, err := http.NewRequest("POST", loginUrl, strings.NewReader(loginBody))
	if err != nil {
		util.Log.Errorf("generate insightcj login post request failed, error %+v!", err)
		return nil
	}
	er.SetHeader((&util.Header{}).InitFromMap(map[string]string{
		"accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Host":         "insightcj.com",
		"Origin":       "https://insightcj.com",
		"Referer":      "https://insightcj.com/signin",
		"Content-Type": "application/x-www-form-urlencoded",
		"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 Edg/85.0.564.63",
	}))
	resp = er.SetRequest(loginPostReq).Do()
	if resp == nil {
		util.Log.Errorf("do insightcj login post request failed!")
		return nil
	}
	return er
}

func FormatResponse(title string) string {
	tmpRuneTxtResponse := []rune(txtResponse)
	var responseRune []rune
	for i := 0; i < 40; i++ {
		index := rand.Int() % len(tmpRuneTxtResponse)
		responseRune = append(responseRune, tmpRuneTxtResponse[index])
	}
	return string(responseRune)
}

func ParseFirstUselessTopic(html string) (url, title string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		util.Log.Errorf("Parse first useless topic from [[%s]] failed, error %+v!", html, err)
		return
	}
	doc.Find(".topic_title").Each(func(i int, s *goquery.Selection) {
		tmpUrl, _ := s.Attr("href")
		tmpTitle, _ := s.Attr("title")
		if url == "" && tmpUrl != "" && strings.HasPrefix(tmpTitle, "灌水专用贴") {
			url = tmpUrl
			title = tmpTitle
		}
	})
	return
}

func ResponseTpoic() bool {
	home := "https://insightcj.com"
	homeGetReq, err := http.NewRequest("GET", home, nil)
	if err != nil {
		util.Log.Errorf("generate insightcj home request failed, error %+v!", err)
		return false
	}
	er.SetHeader((&util.Header{}).InitFromMap(map[string]string{
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Host":       "insightcj.com",
		"Origin":     "https://insightcj.com",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 Edg/85.0.564.63",
	}))
	resp := er.SetRequest(homeGetReq).Do()
	if resp == nil {
		util.Log.Errorf("do insightcj home request failed, error %+v!", err)
		return false
	}
	topic, title := ParseFirstUselessTopic(string(resp))
	if topic == "" || title == "" {
		util.Log.Warnf("can not find first useless topic from insightcj homepage [[%s]]!", string(resp))
		return false
	}
	topicGetReq, err := http.NewRequest("GET", home+topic, nil)
	if err != nil {
		util.Log.Errorf("generate insightcj topic get request failed, error %+v!", err)
		return false
	}
	er.SetHeader((&util.Header{}).InitFromMap(map[string]string{
		"accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Host":       "insightcj.com",
		"Referer":    "https://insightcj.com",
		"user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 Edg/85.0.564.63",
	}))
	resp = er.SetRequest(topicGetReq).Do()
	if resp == nil {
		util.Log.Errorf("do insightcj topic get request failed, error %+v!", err)
		return false
	}
	csrfName, csrfToken, ok := parseCSRF(string(resp))
	if !ok {
		util.Log.Errorf("parse csrf from insightcj topic html failed!")
		return false
	}
	for i := 0; i < maxResponseCount; i++ {
		response := FormatResponse(title)
		if response == "" {
			util.Log.Errorf("format response for [[%s]] failed!", title)
			continue
		}
		responseContent := fmt.Sprintf("r_content=%s&%s=%s", url.QueryEscape(response), url.QueryEscape(csrfName), url.QueryEscape(csrfToken))
		loginPostReq, err := http.NewRequest("POST", home+strings.ReplaceAll(topic, "/topic", "")+"/reply", strings.NewReader(responseContent))
		if err != nil {
			util.Log.Errorf("generate insightcj topic response failed, error %+v!", err)
			continue
		}
		er.SetHeader((&util.Header{}).InitFromMap(map[string]string{
			"accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"Host":         "insightcj.com",
			"Origin":       "https://insightcj.com",
			"Referer":      home + topic,
			"Content-Type": "application/x-www-form-urlencoded",
			"user-agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36 Edg/85.0.564.63",
		}))
		er.SetRequest(loginPostReq).Do()
	}
	return true
}

var er *util.EasyRequest

func main() {
	er = login()
	ResponseTpoic()
}
