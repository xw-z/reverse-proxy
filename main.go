package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

var (
	USER_HOME_DIR, _   = os.UserHomeDir()
	DOCKER_CONIFG_PATH = filepath.Join(USER_HOME_DIR, ".docker")
	CA_FILE            = filepath.Join(DOCKER_CONIFG_PATH, "ca.pem")
	CERT_FILE          = filepath.Join(DOCKER_CONIFG_PATH, "server-cert.pem")
	KEY_FILE           = filepath.Join(DOCKER_CONIFG_PATH, "server-key.pem")
)

var (
	httpAddr   = flag.String("http_addr", ":80", "")
	httpsAddr  = flag.String("https_addr", "", "")
	target     = flag.String("target", "", "--target=http://example.com\n--target=tcp://127.0.0.1:80\n--target=unix:///var/run/docker.sock")
	caFile     = flag.String("tls_ca", CA_FILE, "")
	certFile   = flag.String("tls_cert", CERT_FILE, "")
	keyFile    = flag.String("tls_key", KEY_FILE, "")
	clientAuth = flag.Bool("tls_client_auth", true, "")
)

func ListenHTTP(addr string, handler http.Handler) error {
	s := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Println("http on", addr)
	return s.ListenAndServe()
}

func ListenHTTPS(addr string, caFile, certFile, keyFile string, handler http.Handler) error {
	var clientCAs *x509.CertPool
	if len(caFile) > 0 {
		clientCAs = x509.NewCertPool()
		b, err := ioutil.ReadFile(caFile)
		if err != nil {
			return err
		}
		if ok := clientCAs.AppendCertsFromPEM(b); !ok {
			return fmt.Errorf("ca [%v] invalid", caFile)
		}
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	_ = cert

	tlsConfig := &tls.Config{
		ClientCAs:    clientCAs,
		Certificates: []tls.Certificate{cert},
	}

	if *clientAuth {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	s := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
	log.Println("https on", addr)
	return s.ListenAndServeTLS("", "")
}

func parseNetAddr(s string) (network string, addr string, err error) {
	p := strings.Index(s, "://")
	if p < 0 {
		network = "tcp"
		addr = s
	} else {
		network = s[:p]
		addr = s[p+len("://"):]
	}
	switch network {
	case "http", "tcp", "unix":
	default:
		err = errors.New("unsupport addr " + s)
	}
	return
}

func DispatcherHandler(target string) http.Handler {
	vnetwork, vaddr, err := parseNetAddr(target)
	if err != nil {
		log.Fatalln(err)
	}

	var proxy http.Handler
	if vnetwork == "http" {
		targetUrl, err := url.Parse(target)
		if err != nil {
			log.Fatalln(err)
		}
		proxy = httputil.NewSingleHostReverseProxy(targetUrl)
	} else {
		proxy = &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				r.URL.Scheme = "http"
				r.URL.Host = "backend"
			},
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial(vnetwork, vaddr)
				},
			},
		}
	}
	return proxy
}

func main() {
	flag.Parse()
	if len(*target) <= 0 {
		flag.Usage()
		os.Exit(-1)
	}

	var g errgroup.Group
	var h = DispatcherHandler(*target)

	if len(*httpAddr) > 0 {
		g.Go(func() error {
			log.Fatalln(ListenHTTP(*httpAddr, h))
			return nil
		})
	}
	if len(*httpsAddr) > 0 {
		g.Go(func() error {
			log.Fatalln(ListenHTTPS(*httpsAddr, *caFile, *certFile, *keyFile, h))
			return nil
		})
	}
	g.Wait()
}
