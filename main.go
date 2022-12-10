package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server struct {
	Address string
	Proxy   *httputil.ReverseProxy
	isAlive bool
}

type server interface {
	getAddress() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type Loadbalancer struct {
	Port    string
	Servers []server
	Count   int
}

func newLoadbalancer(servers []server, port string) *Loadbalancer {
	return &Loadbalancer{
		Port:    port,
		Servers: servers,
		Count:   0,
	}

}

func newServer(addr string) *Server {
	url, err := url.Parse(addr)
	if err != nil {
		log.Println("Unable to create new server " + err.Error())
		os.Exit(1)
	}
	return &Server{
		Address: addr,
		Proxy:   httputil.NewSingleHostReverseProxy(url),
		isAlive: true,
	}
}

func (s *Server) getAddress() string { return s.Address }
func (s *Server) IsAlive() bool      { return s.isAlive }

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}
func (lb *Loadbalancer) getAvailableSever() server {
	availServer := lb.Servers[lb.Count%len(lb.Servers)]
	for !availServer.IsAlive() {
		lb.Count++
		availServer = lb.Servers[lb.Count%len(lb.Servers)]
	}
	lb.Count++
	return availServer
}

func (lb *Loadbalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	tarServer := lb.getAvailableSever()
	fmt.Printf("Forwarding request to addrs:: %s \n", tarServer.getAddress())
	tarServer.Serve(w, r)
}
func main() {
	servers := []server{
		newServer("https://www.facebook.com"),
		newServer("https://www.yahoo.com"),
		newServer("https://www.duckduckgo.com"),
	}

	lb := newLoadbalancer(servers, ":8080")

	handlefunc := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handlefunc)

	fmt.Printf("Serving request at Localhost port: %s", lb.Port)
	http.ListenAndServe(":8080", nil)
}
