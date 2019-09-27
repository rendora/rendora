/*
Copyright 2018 George Badawi.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/mafredri/cdp/devtool"
)

// HeadlessClient contains the info of the headless client, most importantly the cdp.Client
type HeadlessClient struct {
	Ctx context.Context
	Cfg *HeadlessConfig
}

//HeadlessResponse contains the status code, DOM content and headers of the response coming from the headless chrome instance
type HeadlessResponse struct {
	Status  int64                  `json:"status"`
	Content string                 `json:"content"`
	Headers map[string]interface{} `json:"headers"`
	Latency float64                `json:"latency"`
}

// HeadlessConfig headless's config
type HeadlessConfig struct {
	UserAgent     string
	Mode          string
	URL           string
	AuthToken     string
	BlockedURLs   []string
	Timeout       int64
	InternalURL   string
	WaitReadyNode string
}

func resolveURLHostname(arg string) (string, error) {
	devURL, err := url.Parse(arg)
	if err != nil {
		return "", err
	}

	devIPs, err := net.LookupIP(devURL.Hostname())

	var devToolURL string
	if err != nil {
		return "", err
	}
	for _, ip := range devIPs {
		devToolURL = ip.String()
	}

	if devURL.Port() == "" {
		devURL.Host = devToolURL
	} else {
		devURL.Host = devToolURL + ":" + devURL.Port()
	}

	return devURL.String(), nil
}

func checkHeadless(arg string) error {
	doCheck := func() error {
		resp, err := http.Get(arg + "/json/version")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}

	for i := 0; i < 4; i++ {
		err := doCheck()
		if err == nil {
			return nil
		}
		log.Println("cannot connect to the headless Chrome instance, trying again after 2 seconds...")
		time.Sleep(2 * time.Second)
	}
	err := doCheck()
	if err == nil {
		return nil
	}
	return errors.New("cannot connect to the headless Chrome instance, make sure it is running")

}

//NewHeadlessClient creates HeadlessClient
func NewHeadlessClient(cfg *HeadlessConfig) (*HeadlessClient, error) {
	ret := &HeadlessClient{
		Cfg: cfg,
	}

	err := checkHeadless(cfg.InternalURL)
	if err != nil {
		return nil, err
	}

	// looks like cdp doesn't resolve hostnames automatically, may lead to problems when used with container networks
	resolvedURL, err := resolveURLHostname(cfg.InternalURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	devTools := devtool.New(resolvedURL)
	pt, err := devTools.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devTools.Create(ctx)
		if err != nil {
			return nil, err
		}
	}

	allocCtx, _ := chromedp.NewRemoteAllocator(ctx, pt.WebSocketDebuggerURL)

	// create chrome instance
	taskCtx, _ := chromedp.NewContext(
		allocCtx,
	)

	ret.Ctx = taskCtx

	return ret, nil
}

// Close close connection
func (c *HeadlessClient) Close() error {
	c.Ctx.Done()
	return nil
}

// GetResponse GoTo navigates to the url, fetches the DOM and returns HeadlessResponse
func (c *HeadlessClient) GetResponse(uri string) (*HeadlessResponse, error) {
	var res string
	timeoutCtx, cancel := context.WithTimeout(c.Ctx, time.Duration(c.Cfg.Timeout)*time.Second)
	defer cancel()

	err := chromedp.Run(timeoutCtx, c.scrapIt(uri, &res))
	if err != nil {
		return nil, err
	}

	return &HeadlessResponse{Content: res}, nil
}

func (c *HeadlessClient) scrapIt(url string, str *string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Emulate(device.Info{UserAgent: c.Cfg.UserAgent}),
		chromedp.Navigate(url),
		chromedp.Sleep(2 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			*str, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	}
}
