package proxy_pool

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"notion-bully/config"
	"regexp"
	"sync"
	"time"
)

type ProxyPool struct {
	sync.RWMutex
	Pool []Proxy
}

type Proxy struct {
	Ip   string
	Port string
	Available bool
	Proxy ProxyInterface
}

type ProxyInterface interface {
	Check(timeoutSeconds int, cancel context.CancelFunc, ch chan string) bool
}

type FreeProxy struct {
	Proxy
}

type ZhimaProxy struct {
	Proxy
	// ip:port
	AssembledString string
}

func Crawler(config config.Config) (string, error) {
	registry := NewRegistry()
	var allProxies []Proxy
	switch config.Proxy.Mode {
	case "freeproxy":
		for _, i := range config.Proxy.FreeProxy.ProxyList {
			registry.Register(i.Url, func(body io.Reader) ([]Proxy, error) {
				bs, err := ioutil.ReadAll(body)
				if err != nil {
					return nil, err
				}
				var proxies []Proxy
				r := exractTable(string(bs))
				for _, j := range r {
					p := Proxy{
						Ip:   j[i.IpColumn],
						Port: j[i.PortColumn],
					}
					proxies = append(proxies, p)
				}
				return proxies, nil
			})
			resp, err := http.Get(i.Url)
			if err != nil {
				return "", err
			}
			ps, err := registry.GetMatch(i.Url)(resp.Body)
			allProxies = append(allProxies, ps...)
		}
		//fmt.Println(allProxies)
		shuffleMapping := make(map[int]struct{})
		proxyChan := make(chan string, 10)
		for i, _ := range allProxies {
			shuffleMapping[i] = struct{}{}
		}
		ctx, cancel := context.WithCancel(context.Background())
		for i := range shuffleMapping {
			go checkProxy(allProxies[i].Ip, allProxies[i].Port, 10, cancel, proxyChan)
		}
		// default max for 60s ti allow for testing proxy status
		go func() {
			time.Sleep(time.Duration(60) * time.Second)
			cancel()
		}()
		for {
			select {
			case <-ctx.Done():
				proxy := <-proxyChan
				return proxy, nil
			default:
				log.Println("waiting...")
				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	case "zhimaproxy":

	default:

	}

	//log.Println(checkProxy(allProxies[1].Ip, allProxies[1].Port, 5))
	return "", nil
}


func (p *ZhimaProxy) Check(timeoutSeconds int, cancel context.CancelFunc, ch chan string) bool {
	conn, err := net.DialTimeout("tcp", p.AssembledString, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		log.Println(err)
		return false
	}
	defer conn.Close()

	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		fmt.Printf("Timeout error: %s\n", err)
		return false
	}

	if err != nil {
		// Log or report the error here
		fmt.Printf("Error: %s\n", err)
		return false
	}
	ch <- "http://" + p.AssembledString
	cancel()
	return true
}

func checkProxy(ip, port string, timeoutSeconds int, cancel context.CancelFunc, ch chan string) bool {
	conn, err := net.DialTimeout("tcp", ip+":"+port, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		log.Println(err)
		return false
	}
	defer conn.Close()

	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		fmt.Printf("Timeout error: %s\n", err)
		return false
	}

	if err != nil {
		// Log or report the error here
		fmt.Printf("Error: %s\n", err)
		return false
	}
	ch <- "http://" + ip + ":" + port
	cancel()
	return true
}

func exractTable(table string) [][]string {
	var result [][]string
	exp, err := regexp.Compile(`<table.*?>.*</table>`)
	if err != nil {
		return nil
	}
	ss := exp.FindAllString(table, -1)
	if len(ss) == 0 {
		return nil
	}
	exp, err = regexp.Compile(`<tr>(<td>(.*?)</td>)+</tr>`)
	re := exp.FindAllString(ss[0], -1)
	for _, line := range re {
		fmt.Println(line)
		exp := regexp.MustCompile(`<td.*?>(.*?)+</td>`)
		re := exp.FindAllStringSubmatch(line, -1)
		var tmp_line []string
		for _, item := range re {
			tmp_line = append(tmp_line, item[1])
		}
		result = append(result, tmp_line)
	}
	return result
}



func (p *ProxyPool) Push(proxy Proxy) {
	p.Lock()
	p.Pool = append(p.Pool, proxy)
	p.Unlock()
}

func (p *ProxyPool) Remove(p2remove Proxy) {
	for i, proxy := range p.Pool {
		if proxy == p2remove {
			p.Lock()
			p.Pool = append(p.Pool[:i], p.Pool[i+1:]...)
			p.Unlock()
		}
	}
}

type MatchRegistry struct {
	matchMapping map[string]func(body io.Reader) ([]Proxy, error)
}

func NewRegistry() *MatchRegistry {
	var match MatchRegistry
	mapping := make(map[string]func(body io.Reader) ([]Proxy, error))
	match.matchMapping = mapping
	return &match
}

func (m MatchRegistry) Register(s string, f func(body io.Reader) ([]Proxy, error)) {
	m.matchMapping[s] = f
}

func (m MatchRegistry) GetMatch(s string) func(body io.Reader) ([]Proxy, error) {
	if f, ok := m.matchMapping[s]; ok {
		return f
	}
	return nil
}
