package main

import (
	"encoding/json"
	"fmt"
	"github.com/aldor007/purge-api/cache"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	purger := cache.NewPurger()

	if os.Getenv("STRAPI_REDIS") != "" {
		redisDB, err :=  strconv.Atoi(os.Getenv("STRAPI_REDIS_DB"))
		if err != nil {
			panic(err)
		}
		purger.AddCache(cache.NewApolloStrapiRedis(cache.ApolloStrapiConfig{
			RedisEndpoint: os.Getenv("STRAPI_REDIS"),
			RedisDB:   redisDB,
		}))

	}
	purger.AddCache(cache.NewNginx(cache.NginxPurgeConfig{
		PurgeMethod: "PURGE",
		URL:         os.Getenv("NGINX_URL"),
	}))
	cf, err := cache.NewCloudflare(cache.CloudflareConfig{
		APIKey:   os.Getenv("CF_API_KEY"),
		APIEmail: os.Getenv("CF_API_EMAIL"),
		ZoneID:   os.Getenv("CF_ZONE_ID"),
	})
	if err != nil {
		panic(err)
	}
	purger.AddCache(cf)
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	purgeRouter := chi.NewRouter()
	purgeRouter.Use(middleware.Logger)
	purgeRouter.Use(middleware.BasicAuth("purge-api", map[string]string{"api": os.Getenv("API_KEY")}))
	purgeRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		toPurge := r.URL.Query().Get("url")
		if toPurge == "" {
			w.WriteHeader(400)
			return
		}

		err := purger.Purge(r.Context(), toPurge)
		if err != nil{
			log.Println("Err purge", toPurge, err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	})
	purgeRouter.Post("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(400)
			return
		}

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Read error", err)
			w.WriteHeader(400)
			return
		}
		purgeData := make(map[string][]string)
		err = json.Unmarshal(body, &purgeData)
		if err != nil {
			log.Println("Parse error", err)
			w.WriteHeader(400)
			return
		}
		toPurgeUrls, ok:= purgeData["urls"]
		if !ok {
			log.Println("No url key")
			w.WriteHeader(400)
			return
		}
		if len(toPurgeUrls) == 0 {
			log.Println("urls empty")
			w.WriteHeader(400)
			return
		}

		errs := make([]error, 0)
		for _, toPurge := range toPurgeUrls {
			err = purger.Purge(r.Context(), toPurge)
			if err != nil{
				log.Println("Err purge", toPurge, err)
				errs = append(errs, err)
			}

		}
		if len(errs) == 0 {
			w.WriteHeader(202)
		} else {
			w.WriteHeader(500)
			fmt.Fprint(w, errs)
		}
	})
	r.Mount("/purge", purgeRouter)
	srv := &http.Server{
		Addr:         ":8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: r, // Pass our instance of gorilla/mux in.
	}

	log.Println("Server listening :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}