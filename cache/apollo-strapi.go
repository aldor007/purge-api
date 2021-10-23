package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/url"
	"strings"
)


type ApolloStrapiConfig struct {
	RedisEndpoint string
	RedisDB int
}

type apolloExt struct {
	PersistedQuery struct {
		Version    int    `json:"version"`
		Sha256Hash string `json:"sha256Hash"`
	} `json:"persistedQuery"`
}
type ApolloStrapiRedis struct {
	config ApolloStrapiConfig
	client *redis.Client
}

func NewApolloStrapiRedis(config ApolloStrapiConfig) (*ApolloStrapiRedis) {
	client :=  redis.NewClient(&redis.Options{
		Addr:     config.RedisEndpoint ,
		Password: "", // no password set
		DB:       config.RedisDB,  // use default DB
	})

	return &ApolloStrapiRedis{
		client: client,
		config: config,
	}
}

func (a *ApolloStrapiRedis) Purge(ctx context.Context,urlToPurge string) error  {
	if !strings.Contains(urlToPurge, "/graphql") {
		return nil
	}

	decodedValue, err := url.QueryUnescape(urlToPurge)
	if err != nil {
		return err
	}

	p, err := url.Parse(decodedValue)
	if err != nil {
		return err
	}

	exts := p.Query().Get("extensions")
	if exts== "" {
		return nil
	}

	var ext apolloExt
	err = json.Unmarshal([]byte(exts), &ext)
	if err != nil {
		return err
	}
	if ext.PersistedQuery.Sha256Hash == "" {
		return nil
	}
	key := fmt.Sprintf("apq:%s", ext.PersistedQuery.Sha256Hash)
	result := a.client.Del(ctx, key)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}
