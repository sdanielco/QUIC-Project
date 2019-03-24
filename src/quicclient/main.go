package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io"
	"fmt"
	"net/http"
	"sync"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/lucas-clemente/quic-go/h2quic"
)

func main() {
	quiet := flag.Bool("q", false, "don't print the data")
	flag.Parse()


	roundTripper := &h2quic.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs: GetRootCA(),
		},
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}

	urls := []string{"https://localhost:443/index.html"}
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, addr := range urls {
		fmt.Printf("GET %s\n", addr)
		go func(addr string) {
			rsp, err := hclient.Get(addr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Got response for %s: %#v\n", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				panic(err)
			}
			if *quiet {
				fmt.Printf("Request Body: %d bytes\n", body.Len())
			} else {
				fmt.Println("Request Body:")
				fmt.Printf("%s\n", body.Bytes())
			}
			wg.Done()
		}(addr)
	}
	wg.Wait()
}

func GetRootCA() *x509.CertPool {
	caCertPath := "ca.pem"
	caCertRaw, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		panic(err)
	}
	p, _ := pem.Decode(caCertRaw)
	if p.Type != "CERTIFICATE" {
		panic("expected a certificate")
	}
	caCert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)
	return certPool
}
