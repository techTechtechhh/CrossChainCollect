package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"golang.org/x/net/context/ctxhttp"
)

type HttpOption struct {
	Method      string
	Host        string
	Url         *url.URL
	Header      map[string]string
	RequestBody interface{}
	Response    interface{}
	Proxy       string
}

var EtherscanTimeout time.Duration = 40

func (ho *HttpOption) Send(ctx context.Context) error {
	log.Debug("http option send", "method", ho.Method, "url", ho.Url, "proxy", ho.Proxy)
	if ho.Url == nil {
		return fmt.Errorf("no url specificed")
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(ho.RequestBody); err != nil {
		return err
	}
	req, err := http.NewRequest(ho.Method, ho.Url.String(), &buf)
	if err != nil {
		return err
	}

	if ho.Host != "" {
		req.Host = ho.Host
	}
	if ho.Header != nil {
		for k, v := range ho.Header {
			req.Header.Set(k, v)
		}
	}
	client := &http.Client{}
	if ho.Proxy != "" {
		proxy, err := url.Parse(ho.Proxy)
		if err != nil {
			return err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	resp, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, ho.Response); err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("StatusCode: %d", resp.StatusCode)
	}
	return nil
}

func IsHttps(uri string) bool {
	return strings.Index(strings.ToLower(uri), "https://") == 0
}

func IsHttp(uri string) bool {
	return strings.Index(strings.ToLower(uri), "http://") == 0
}

func HttpGet(url string) ([]byte, error) {
	// filler 模块中，并发请求如果不设置超时，导致goroutine阻塞，k8s容器爆内存
	client := &http.Client{
		Timeout: EtherscanTimeout * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func HttpGetWithProxy(uri string, proxyUrls []string) ([]byte, error) {
	index := rand.Intn(len(proxyUrls))
	proxyUrl := proxyUrls[index]
	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: EtherscanTimeout * time.Second,
		Transport: &http.Transport{
			// 设置代理
			Proxy: http.ProxyURL(proxy),
		},
	}
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func HttpGetObjectWithProxy(url string, proxy []string, dest any) error {
	if len(proxy) == 0 {
		return HttpGetObject(url, dest)
	}
	data, err := HttpGetWithProxy(url, proxy)
	//fmt.Println(string(data))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("%v, raw data: %v", err, string(data))
	}
	return nil
}

func HttpGetObject(url string, dest any) error {
	data, err := HttpGet(url)
	if err != nil {
		return err
	}
	// fmt.Println("======")
	// fmt.Println(string(data))
	// fmt.Println("======")

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("%v, raw data: %v", err, string(data))
	}

	return nil
}

func GeneratePhalconLink(chain, hash string) string {
	var c string
	if chain == "fantom" {
		c = "ftm"
	} else if chain == "avalanche" {
		c = "avax"
	} else {
		c = chain
	}
	return (fmt.Sprintf("https://explorer.phalcon.xyz/tx/%s/%s", c, hash))
}

// dns resolve, tcp connect refuse, timeout
func IsNetError(err error) bool {
	if _, ok := err.(net.Error); ok {
		return true
	}
	return false
	// netErr, ok := err.(net.Error)
	// if !ok {
	// 	return false
	// }

	// if netErr.Timeout() {
	// 	return true
	// }

	// opErr, ok := netErr.(*net.OpError)
	// if !ok {
	// 	return false
	// }

	// switch t := opErr.Err.(type) {
	// case *net.DNSError:
	// 	return true
	// case *os.SyscallError:
	// 	if errno, ok := t.Err.(syscall.Errno); ok {
	// 		// switch errno {
	// 		// case syscall.ECONNREFUSED, syscall.ETIMEDOUT:
	// 		// 	return true
	// 		// }
	// 		return errno.Temporary() || errno.Timeout()
	// 	}
	// }
	// return false
}
