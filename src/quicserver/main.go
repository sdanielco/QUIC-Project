package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"net/http"
	"strings"
	"sync"

	"github.com/lucas-clemente/quic-go/h2quic"
)

type binds []string

func (b binds) String() string {
	return strings.Join(b, ",")
}

func (b *binds) Set(v string) error {
	*b = strings.Split(v, ",")
	return nil
}

// Size is needed by the /demo/upload handler to determine the size of the uploaded file
type Size interface {
	Size() int64
}



func main() {
	bs := binds{}
	flag.Var(&bs, "bind", "bind to")
	www := flag.String("www", "C:\\www", "www data")
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*www)))

	if len(bs) == 0 {
		bs = binds{"0.0.0.0:443"}
	}

	var wg sync.WaitGroup
	wg.Add(len(bs))
	for _, b := range bs {
		bCap := b
		go func() {
			var err error
			if *tcp {
				err = h2quic.ListenAndServe(bCap, "cert.pem", "priv.key", nil)
			} else {
				server := h2quic.Server{
					Server: &http.Server{Addr: bCap},
					QuicConfig: &quic.Config{
						StatelessResetKey: bytes.Repeat([]byte("A"), 32),
					},
				}
				err = server.ListenAndServeTLS("cert.pem", "priv.key")
			}
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
