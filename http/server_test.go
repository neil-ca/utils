package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ulicod3/utils/http/handlers"
)

func TestSimpleHttpServer(t *testing.T) {
    srv := &http.Server{
        Addr: "127.0.0.1:8080",
        Handler: http.TimeoutHandler(
            handlers.DefaultHandler(), 2*time.Minute, ""),
            IdleTimeout: 5 * time.Minute,
            ReadHeaderTimeout: time.Minute,
        }

        l, err := net.Listen("tcp", srv.Addr)
        if err != nil { t.Fatal(err) }

        go func() {
            err := srv.Serve(l)
            if err != http.ErrServerClosed {
                t.Error(err)
            }
        }()

        testCases := []struct {
            method   string
            body     io.Reader
            code     int
            response string
        }{
            {http.MethodGet, nil, http.StatusOK, "Hello friend!"},
            {http.MethodPost, bytes.NewBufferString("<world>"), http.StatusOK, "Hello, &lt;world&gt;!"},
            {http.MethodHead, nil, http.StatusMethodNotAllowed, ""},
        }

        client := new(http.Client)
        path := fmt.Sprintf("http://%s/", srv.Addr)
    }


