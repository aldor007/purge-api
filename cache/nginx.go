package cache

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type NginxPurgeConfig struct {
	PurgeMethod string
	URL string
}


type Nginx struct {
	client *http.Client
	config NginxPurgeConfig
	endpoint string
}

func NewNginx(config NginxPurgeConfig) *Nginx {

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
	u, _  := url.Parse(config.URL)

	return &Nginx{config: config, client: client, endpoint:  u.Host}
}

func (n *Nginx) createPurgeRe(ctx context.Context, url string, headers map[string]string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, n.config.PurgeMethod, url, nil)
	if err != nil {
		return nil, err
	}

	req.URL.Host = n.endpoint
	req.URL.Scheme = "http"
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return  req, nil
}
func (n *Nginx) Purge(ctx context.Context, url string) error  {
	purgeReq := make(chan *http.Request, 5)
	errorsChan := make(chan error)
	go func() {
		for req := range purgeReq {
			res, err := n.client.Do(req)
			if err != nil {
				errorsChan <- err
			}

			log.Println("Nginx purge", url, res.StatusCode, req.Header)
			if res.StatusCode != 200 && res.StatusCode != 404 {

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
