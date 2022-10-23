package cache

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"log"
)

type CloudflareConfig struct {
	APIKey string
	APIEmail string
	ZoneID string
}


type Cloudflare struct {
	client *cloudflare.API
	config CloudflareConfig
}

func NewCloudflare(config CloudflareConfig) (*Cloudflare, error) {


	api, err := cloudflare.New(config.APIKey, config.APIEmail)

	if err != nil {
		return nil, err
	}

	return &Cloudflare{config: config, client: api}, nil
}

func (c *Cloudflare) Purge(ctx context.Context,url string) error  {
	pReq := cloudflare.PurgeCacheRequest{
		Everything: true,
		Files:      []string{url},
		Tags:       nil,
		Hosts:      nil,
		Prefixes:  nil,
	}
	res, err := c.client.PurgeCacheContext(ctx, c.config.ZoneID, pReq)
	if err != nil {
		log.Println("CF purge err", err)
		return err
	}
	if res.Success {
		log.Println("CF purge ", url)
	}
	return nil
}
