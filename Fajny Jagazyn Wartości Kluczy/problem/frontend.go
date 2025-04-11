package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type unixDialer struct {
	net.Dialer
}

func (d *unixDialer) Dial(network, address string) (net.Conn, error) {
	return d.Dialer.Dial("unix", "/tmp/kv."+strings.Split(address, ":")[0]+"/kv.socket")
}

var transport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial:  (&unixDialer{net.Dialer{Timeout: 5 * time.Second}}).Dial,
}

var backends sync.Map

func NewKV() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	session := hex.EncodeToString(bytes)

	go func() {
		cmd := exec.Command("./kv")
		cmd.Env = append(os.Environ(), "SESSION="+session)

		cmd.Run()
		backends.Delete(session)
	}()

	url, err := url.Parse("http://" + session)
	if err != nil {
		return ""
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = transport

	backends.Store(session, proxy)
	return session
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		session := ""
		if cookie, err := r.Cookie("session"); err == nil {
			session = cookie.Value
		}

		proxy, ok := backends.Load(session)
		if !ok {
			cookie := &http.Cookie{Name: "session", Value: NewKV(), Path: "/", Expires: time.Now().Add(180 * time.Second)}
			http.SetCookie(w, cookie)
			w.Write([]byte("We booted a fresh web scale Key Value Store just for you 🥰 (Please enjoy it for the next 180 seconds)"))
			return
		}
		proxy.(*httputil.ReverseProxy).ServeHTTP(w, r)
	})

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
		Handler:      http.DefaultServeMux,
		Addr:         ":1024",
	}
	log.Println(srv.ListenAndServe())
}
