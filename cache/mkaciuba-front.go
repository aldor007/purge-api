
package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)



type mkaciubaFront struct {
	client *http.Client
	endpoint string
	apiKey string
}

func NewMKaciubaFront(endpoint string, apiKey string) *mkaciubaFront {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return &mkaciubaFront{apiKey: apiKey, client: client, endpoint:  endpoint}
}

func (n *mkaciubaFront) createPurgeRe(ctx context.Context, url string, headers map[string]string) (*http.Request, error) {
	toPurge := strings.Replace(url, "https://mkaciuba.pl", "", 1)
	if !strings.HasPrefix(toPurge, "/" ) {
		toPurge = "/" + toPurge
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", n.endpoint + "?path=" + toPurge , nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("x-api-key", n.apiKey)

	return  req, nil
}
func (n *mkaciubaFront) Purge(ctx context.Context, url string) error  {
	if !strings.Contains(url, "mkaciuba.pl") {
		return nil
	}

	purgeReq := make(chan *http.Request, 5)
	errorsChan := make(chan error)
	go func() {
		for req := range purgeReq {
			res, err := n.client.Do(req)
			if err != nil {
				errorsChan <- err
			}

			log.Println("mkaciubaFront purge", url, res.StatusCode, req.Header, req.URL.String())
			if res.StatusCode != 201 && res.StatusCode != 404 {

				errorsChan <- errors.New(fmt.Sprintf("http error %d", res.StatusCode))
			}
		}
		close(errorsChan)
	}()


	req, err := n.createPurgeRe(ctx, url, map[string]string{"x-forwarded-proto": "https"})
	if err != nil {
		return err
	}
	purgeReq <- req
	req, err = n.createPurgeRe(ctx, url, map[string]string{"x-forwarded-proto": "https", "accept-encoding": "gzip, br"})
	if err != nil {
		return err
	}
	purgeReq <- req
	req, err = n.createPurgeRe(ctx, url, map[string]string{"x-forwarded-proto": "https", "accept-encoding": "gzip"})
	if err != nil {
		return err
	}
	purgeReq <- req
	req, err = n.createPurgeRe(ctx, url, map[string]string{"x-forwarded-proto": "https", "accept-encoding": "gzip", "x-origin-method": "GET"})
	if err != nil {
		return err
	}
	purgeReq <- req
	req, err = n.createPurgeRe(ctx, url, map[string]string{"x-forwarded-proto": "https", "accept-encoding": "br", "x-origin-method": "GET"})
	if err != nil {
		return err
	}
	purgeReq <- req
	close(purgeReq)


	var mainErr error
	for errP := range errorsChan {
		mainErr = errP
	}

	return mainErr
}
