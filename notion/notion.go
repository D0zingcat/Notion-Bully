package notion

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	proxy_pool "notion-bully/proxy-pool"
	"notion-bully/util"
	"strings"
	"time"
)

type NotionBusiness interface {
	VisitReferPage()
	ValidateMail(ul string)
	SendCode(ul string) string
	Login(ul string, csrf, pass string) string
	ActivateRefer(ul string)
	SendRefmail() bool
	SubmitTransaction(ul, userId, timezone, firstname, lastname, src, job, scope string)
}

type NotionJob struct {
	Client    http.Client
	Mail      string
	ReferLink string
	UserAgent string
}

func NewNotionJob(userAgents []string, refer string, proxy string) *NotionJob {
	var notion NotionJob
	notion.UserAgent = util.GetOneRandom(userAgents).(string)
	cookieJar, _ := cookiejar.New(nil)
	var transport *http.Transport
	if len(proxy) == 0 {
		transport = &http.Transport{}
	} else {
		ul, err := url.Parse(proxy)
		if err != nil {
			log.Println("fail to parse proxy url", err)
		} else {
			transport = &http.Transport{
				Proxy: http.ProxyURL(ul),
			}
		}

	}
	notion.Client = http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           cookieJar,
		Timeout:       0,
	}
	notion.ReferLink = refer
	return &notion
}

func (n *NotionJob) SetProxy(p proxy_pool.Proxy) {

}

func (n *NotionJob) VisitReferPage() {
	req, _ := http.NewRequest("GET", n.ReferLink, nil)
	req.Header.Set(`authority`, `www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`referer`, `https://www.notion.so/`)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ := httputil.DumpResponse(resp, true)
	_ = dump
	//log.Println(string(dump))
}

func (n *NotionJob) ValidateMail(ul, mail string) {
	// validate mail
	req, _ := http.NewRequest("POST", ul, strings.NewReader(fmt.Sprintf(`{"email":"%s"}`, mail)))
	req.Header.Set(`authority`, `www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`referer`, n.ReferLink)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ := httputil.DumpResponse(resp, true)
	log.Println(string(dump))
}

func (n *NotionJob) SendCode(ul, mail string) string {
	req, _ := http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"email":"%s","disableLoginLink":false,"native":false,"isSignup":true}`, mail)))
	req.Header.Set(`authority`, `www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`referer`, n.ReferLink)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	type ret struct {
		CsrfState string `json:"csrfState"`
	}
	var re ret
	json.NewDecoder(resp.Body).Decode(&re)
	return re.CsrfState
}

func (n *NotionJob) Login(ul, state, pass string) string {
	req, _ := http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"state":"%s","password":"%s"}`, state, pass)))
	req.Header.Set(`authority`, `www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`Origin`, `https://www.notion.so`)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ := httputil.DumpRequest(req, true)
	log.Println(string(dump))
	rebytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	userId := util.GetOneLevelJson(string(rebytes), "userId")
	dump, _ = httputil.DumpResponse(resp, true)
	log.Println(string(dump))
	return userId
}

func (n *NotionJob) SubmitTransaction(ul string, userId, timezone, firstname, lastname, src, job, scope string) {
	req, _ := http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"operations":[{"id":"%s","table":"user_settings","path":["settings"],"command":"update","args":{"locale":"en","time_zone":"%s","used_desktop_web_app":true}}]}`, userId, timezone)))
	req.Header.Set(`origin`, `https://www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`referer`, `https://www.notion.so/onboarding`)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ := httputil.DumpRequest(req, true)
	log.Println(string(dump))
	dump, _ = httputil.DumpResponse(resp, true)
	log.Println(string(dump))

	req, _ = http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"operations":[{"id":"%s","table":"notion_user","path":[],"command":"update","args":{"given_name":"%s","family_name":"%s"}},{"id":"%s","table":"user_settings","path":["settings"],"command":"update","args":{"persona":"%s","type":"%s","source":"%s"}}]}`,
		userId, firstname, lastname, userId, job, scope, src)))
	req.Header.Set(`origin`, `https://www.notion.so`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`referer`, `https://www.notion.so/onboarding`)
	resp, err = n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ = httputil.DumpRequest(req, true)
	log.Println(string(dump))
	dump, _ = httputil.DumpResponse(resp, true)
	log.Println(string(dump))
	// submit this user has been registered
	go func() {
		req, _ = http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"operations":[{"id":"%s","table":"notion_user","path":[],"command":"update","args":{"onboarding_completed":true}}]}`,
			userId)))
		req.Header.Set(`origin`, `https://www.notion.so`)
		req.Header.Set(`content-type`, `application/json`)
		req.Header.Set(`user-agent`, n.UserAgent)
		req.Header.Set(`referer`, `https://www.notion.so/onboarding`)
		resp, err = n.Client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		dump, _ = httputil.DumpRequest(req, true)
		log.Println(string(dump))
		dump, _ = httputil.DumpResponse(resp, true)
		log.Println(string(dump))
	}()
}

func (n *NotionJob) ActivateRefer(ul string) {
	u, err := url.Parse(n.ReferLink)
	if err != nil {
		log.Fatal("fail to query refer code")
	}
	refUid := u.Query().Get("r")
	refUid = refUid[:8] + "-" + refUid[8:12] + "-" + refUid[12:16] + "-" + refUid[16:20] + "-" + refUid[20:]
	log.Println(refUid)
	req, _ := http.NewRequest(`POST`, ul, strings.NewReader(fmt.Sprintf(`{"fromUserId":"%s"}`, refUid)))
	req.Header.Set(`Referer`, `https://www.notion.so/onboarding`)
	req.Header.Set(`content-type`, `application/json`)
	req.Header.Set(`user-agent`, n.UserAgent)
	req.Header.Set(`Origin`, `https://www.notion.so`)
	resp, err := n.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	dump, _ := httputil.DumpRequest(req, true)
	log.Println(string(dump))
	dump, _ = httputil.DumpResponse(resp, true)
	log.Println(string(dump))
	// sleep for 10s to wait for the last transaction to finished
	time.Sleep(time.Duration(10) * time.Second)
}

func (n NotionJob) SendRefmail() bool {
	return false
}
