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

package rendora

import (
	"context"
	"errors"
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

var defaultBlockedURLs []string

//headlessClient contains the info of the headless client, most importantly the cdp.Client
type headlessClient struct {
	RPCConn *rpcc.Conn
	C       *cdp.Client
	Mtx     *sync.Mutex
	rendora *Rendora
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

func checkHeadless(arg string, logsMode string) error {
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
		if logsMode != "NONE" {
			log.Println("Cannot connect to the headless Chrome instance, trying again after 2 seconds...")
		}
		time.Sleep(2 * time.Second)
	}
	err := doCheck()
	if err == nil {
		return nil
	}
	return errors.New("Cannot connect to the headless Chrome instance, make sure it is running")

}

//NewHeadlessClient creates HeadlessClient
func (R *Rendora) newHeadlessClient() error {
	ret := &headlessClient{
		Mtx:     &sync.Mutex{},
		rendora: R,
	}
	ctx := context.Background()

	err := checkHeadless(R.c.Headless.Internal.URL, R.c.LogsMode)
	if err != nil {
		return err
	}

	// looks like cdp doesn't resolve hostnames automatically, may lead to problems when used with container networks
	resolvedURL, err := resolveURLHostname(R.c.Headless.Internal.URL)
	if err != nil {
		return err
	}

	devt := devtool.New(resolvedURL)
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return err
		}
	}

	ret.RPCConn, err = rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}

	ret.C = cdp.NewClient(ret.RPCConn)

	domContent, err := ret.C.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()

	if err = ret.C.Page.Enable(ctx); err != nil {
		return err
	}

	err = ret.C.Network.Enable(ctx, nil)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-Rendora-Type": "RENDER",
	}

	headersStr, err := json.Marshal(headers)
	if err != nil {
		return err
	}

	err = ret.C.Network.SetExtraHTTPHeaders(ctx, network.NewSetExtraHTTPHeadersArgs(headersStr))
	if err != nil {
		return err
	}

	err = ret.C.Network.SetCacheDisabled(ctx, network.NewSetCacheDisabledArgs(R.c.Headless.CacheDisabled))
	if err != nil {
		return err
	}

	blockedURLs := network.NewSetBlockedURLsArgs(defaultBlockedURLs)

	err = ret.C.Network.SetBlockedURLs(ctx, blockedURLs)
	if err != nil {
		return err
	}

	R.h = ret

	return nil
}

//GoTo navigates to the url, fetches the DOM and returns HeadlessResponse
func (c *headlessClient) getResponse(uri string) (*HeadlessResponse, error) {

	c.Mtx.Lock()
	defer c.Mtx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.rendora.c.Headless.Timeout)*time.Second)
	defer cancel()

	timeStart := time.Now()
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

	waitUntil := c.rendora.c.Headless.WaitAfterDOMLoad
	if waitUntil > 0 {
		time.Sleep(time.Duration(waitUntil) * time.Millisecond)
	}

	if _, err = domContent.Recv(); err != nil {
		return nil, err
	}

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

	elapsed := float64(time.Since(timeStart)) / float64(time.Duration(1*time.Millisecond))

	if c.rendora.c.Server.Enable {
		c.rendora.metrics.Duration.Observe(elapsed)
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
		Latency: elapsed,
	}

	return ret, nil
}
