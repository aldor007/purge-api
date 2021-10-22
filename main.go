package main

import (
	"io/ioutil"
	"log"
	"github.com/aldor007/purge-api/cache"
	"net/http"
	"os"
	"time"
	"encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	purger := cache.NewPurger()
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
	r.Use(middleware.Logger)
	r.Use(middleware.BasicAuth("purge-api", map[string]string{"api": os.Getenv("API_KEY")}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/purge", func(w http.ResponseWriter, r *http.Request) {
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
	r.Post("/purge", func(w http.ResponseWriter, r *http.Request) {
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
		purgeData := make(map[string]string)
		err = json.Unmarshal(body, &purgeData)
		if err != nil {
			log.Println("Parse error", err)
			w.WriteHeader(400)
			return
		}
		toPurge, ok:= purgeData["url"]
		if !ok {
			log.Println("No url key")
			w.WriteHeader(400)
			return
		}
		if toPurge == "" {
			log.Println("url empty")
			w.WriteHeader(400)
			return
		}

		err = purger.Purge(r.Context(), toPurge)
		if err != nil{
			log.Println("Err purge", toPurge, err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	})
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