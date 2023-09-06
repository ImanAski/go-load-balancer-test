package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type IServer interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}
type Server struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []IServer
}

func newServer(addr string) *Server {
	serverUrl, err := url.Parse(addr)
	handleError(err)

	return &Server{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func newLoadBalancer(port string, servers []IServer) *LoadBalancer {

	return &LoadBalancer{
		port:            port,
		servers:         servers,
		roundRobinCount: 0,
	}

}

func (s *Server) Address() string {
  return s.addr
}

func (s *Server) IsAlive() bool {
  return true
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request){
  s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextServer() IServer{
  server := lb.servers[lb.roundRobinCount%len(lb.servers)]
  for !server.IsAlive(){
    lb.roundRobinCount++
    server = lb.servers[lb.roundRobinCount%len(lb.servers)]
  }
  lb.roundRobinCount++
  return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
  targetServer := lb.getNextServer()
  fmt.Printf("forwarding request to address %q\n", targetServer.Address())
  targetServer.Serve(w, r)

}

func handleError(err error) {
	if err != nil {
		fmt.Printf("There was an error -> %v\n", err)
		os.Exit(1)
	}
}

func main() {
  servers := []IServer{
    newServer("https://www.google.com"),
    newServer("https://www.bing.com"),
    newServer("https://www.yahoo.com"),
  }

  lb := newLoadBalancer("8000", servers)

  handleRedirect := func(w http.ResponseWriter, r *http.Request){
    lb.serveProxy(w, r)
  }
  http.HandleFunc("/", handleRedirect)

  fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
  http.ListenAndServe(":"+lb.port, nil)
}
