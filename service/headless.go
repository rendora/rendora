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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
)

// HeadlessClient contains the info of the headless client, most importantly the cdp.Client
type HeadlessClient struct {
	RPCConn *rpcc.Conn
	C       *cdp.Client
	Mtx     *sync.Mutex
	Cfg     *HeadlessConfig
}

//HeadlessResponse contains the status code, DOM content and headers of the response coming from the headless chrome instance
type HeadlessResponse struct {
	Status  int               `json:"status"`
	Content string            `json:"content"`
	Headers map[string]string `json:"headers"`
	Latency float64           `json:"latency"`
}

// HeadlessConfig headless's config
type HeadlessConfig struct {
	Mode        string
	URL         string
	AuthToken   string
	BlockedURLs []string
	Timeout     uint16
	InternalURL string
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
		log.Println("Cannot connect to the headless Chrome instance, trying again after 2 seconds...")
		time.Sleep(2 * time.Second)
	}
	err := doCheck()
	if err == nil {
		return nil
	}
	return errors.New("Cannot connect to the headless Chrome instance, make sure it is running")

}

//NewHeadlessClient creates HeadlessClient
func NewHeadlessClient(cfg *HeadlessConfig) (*HeadlessClient, error) {
	ret := &HeadlessClient{
		Mtx: &sync.Mutex{},
		Cfg: cfg,
	}
	ctx := context.Background()

	err := checkHeadless(cfg.InternalURL)
	if err != nil {
		return nil, err
	}

	// looks like cdp doesn't resolve hostnames automatically, may lead to problems when used with container networks
	resolvedURL, err := resolveURLHostname(cfg.InternalURL)
	if err != nil {
		return nil, err
	}

	devt := devtool.New(resolvedURL)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return nil, err
		}
	}

	ret.RPCConn, err = rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}

	ret.C = cdp.NewClient(ret.RPCConn)

	domContent, err := ret.C.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer domContent.Close()

	if err = ret.C.Page.Enable(ctx); err != nil {
		return nil, err
	}

	err = ret.C.Network.Enable(ctx, nil)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"X-Rendora-Type": "RENDER",
	}

	headersStr, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	err = ret.C.Network.SetExtraHTTPHeaders(ctx, network.NewSetExtraHTTPHeadersArgs(headersStr))
	if err != nil {
		return nil, err
	}

	blockedURLs := network.NewSetBlockedURLsArgs(cfg.BlockedURLs)

	err = ret.C.Network.SetBlockedURLs(ctx, blockedURLs)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

// GetResponse GoTo navigates to the url, fetches the DOM and returns HeadlessResponse
func (c *HeadlessClient) GetResponse(uri string) (*HeadlessResponse, error) {
	c.Mtx.Lock()
	defer c.Mtx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Cfg.Timeout)*time.Second)
	defer cancel()

	navArgs := page.NewNavigateArgs(uri)
	networkResponse, err := c.C.Network.ResponseReceived(ctx)
	if err != nil {
		return nil, err
	}

	_, err = c.C.Page.Navigate(ctx, navArgs)
	if err != nil {
		return nil, err
	}

	responseReply, err := networkResponse.Recv()
	if err != nil {
		return nil, err
	}

	domContent, err := c.C.Page.DOMContentEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer domContent.Close()

	loadEventFired, err := c.C.Page.LoadEventFired(ctx)
	if err != nil {
		return nil, err
	}
	defer loadEventFired.Close()

	for {
		select {
		case <-domContent.Ready():
			if _, err = domContent.Recv(); err != nil {
				return nil, err
			}
		case <-loadEventFired.Ready():
			doc, err := c.C.DOM.GetDocument(ctx, nil)
			if err != nil {
				return nil, err
			}

			domResponse, err := c.C.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
				NodeID: &doc.Root.NodeID,
			})
			if err != nil {
				return nil, err
			}

			responseHeaders := make(map[string]string)
			err = json.Unmarshal(responseReply.Response.Headers, &responseHeaders)
			if err != nil {
				return nil, err
			}

			ret := &HeadlessResponse{
				Content: domResponse.OuterHTML,
				Status:  responseReply.Response.Status,
				Headers: responseHeaders,
			}

			return ret, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("reponse timeout from headless chrome")
		}
	}
}
