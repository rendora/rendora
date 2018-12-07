package main

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

//HeadlessClient contains the info of the headless client, most importantly the cdp.Client
type HeadlessClient struct {
	RPCConn *rpcc.Conn
	C       *cdp.Client
	Mtx     *sync.Mutex
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
func NewHeadlessClient() (*HeadlessClient, error) {
	ret := &HeadlessClient{
		Mtx: &sync.Mutex{},
	}
	ctx := context.Background()

	err := checkHeadless(Rendora.C.Headless.Internal.URL)
	if err != nil {
		return nil, err
	}

	// looks like cdp doesn't resolve hostnames automatically, may lead to problems when used with container networks
	resolvedURL, err := resolveURLHostname(Rendora.C.Headless.Internal.URL)
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

	blockedURLs := network.NewSetBlockedURLsArgs(defaultBlockedURLs)

	err = ret.C.Network.SetBlockedURLs(ctx, blockedURLs)

	if err != nil {
		return nil, err
	}

	return ret, nil
}

//GoTo navigates to the url, fetches the DOM and returns HeadlessResponse
func (c *HeadlessClient) GoTo(uri string) (*HeadlessResponse, error) {

	c.Mtx.Lock()
	defer c.Mtx.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

	waitUntil := Rendora.C.Headless.WaitAfterDOMLoad
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

	if Rendora.C.Server.Enable {
		elapsed := float64(time.Since(timeStart)) / float64(time.Duration(1*time.Millisecond))
		Rendora.M.Duration.Observe(elapsed)
	}

	responseHeaders := make(map[string]string)
	err = json.Unmarshal(responseReply.Response.Headers, &responseHeaders)
	if err != nil {
		return nil, err
	}
	ret := &HeadlessResponse{
		Body:    domResponse.OuterHTML,
		Status:  responseReply.Response.Status,
		Headers: responseHeaders,
	}

	return ret, nil
}
