# cache-api

Simple application that will purge cache from nginx (using https://github.com/DreamLab/ngx_cache_purge), Cloudflare and 
strapi apollo persistent queries from redis

# Usage

Run server
```bash
$ API_KEY=foqwe NGINX_URL=http://10.8.0.50 CF_API_KEY=Key CF_API_EMAIL=something@at.com CF_ZONE_ID=zoneId go run main.go                                                                                                                     
```

Call api

```bash 
curl -X POST -u api:foqwe  http://172.23.62.218:8080/purge -d '{ "url": "https://mkaciuba.pl/graphql?operationName=categoryBySlug&variables=%7B%22categorySlug%22%3A%22milena-studio-2021%22%2C%22limit%22%3A20%2C%22start%22%3A20%7D&extensions=%7B%22persistedQuery%22%3A%7B%22version%22%3A1%2C%22sha256Hash%22%3A%22cfdbf11f396b8da29b65008239b344fc8eda1f74a402e1031b332a9e643f0b95%22%7D%7D"}'
```

