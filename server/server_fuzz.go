// +build gofuzz

package server

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"strings"
)

// FuzzProfile tests the profile server.
func FuzzProfile(fuzz []byte) int {
	if len(fuzz) == 0 {
		return -1
	}
	server, err := New(EnableProfiler())
	if err != nil {
		panic(err.Error())
	}
	server.Run()
	defer server.profileListener.Close()
	defer server.httpListener.Close()
	address := server.profileListener.Addr
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err.Error())
	}
	_, err = connection.Write(fuzz)
	if err != nil {
		return 0
	}
	response, err := ioutil.ReadAll(connection)
	if err != nil {
		return 0
	}
	err = connection.Close()
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(response))
	if len(fields) < 2 {
		panic("invalid HTTP response")
	}
	code := fields[1]
	if code == "500" {
		panic("server panicked")
	}
	return 1
}

// FuzzHTTP tests the HTTP server.
func FuzzHTTP(fuzz []byte) int {
	if len(fuzz) == 0 {
		return -1
	}
	server, err := New(Listener("localhost"))
	if err != nil {
		panic(err.Error())
	}
	server.Run()
	defer server.httpListener.Close()
	address := server.httpListener.Addr
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err.Error())
	}
	_, err = connection.Write(fuzz)
	if err != nil {
		return 0
	}
	response, err := ioutil.ReadAll(connection)
	if err != nil {
		return 0
	}
	err = connection.Close()
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(response))
	if len(fields) < 2 {
		panic("invalid HTTP response")
	}
	code := fields[1]
	if code == "500" {
		panic("server panicked")
	}
	return 1
}

// FuzzHTTPS tests the HTTPS server.
func FuzzHTTPS(fuzz []byte) int {
	if len(fuzz) == 0 {
		return -1
	}
	server, err := New(TLSListener("localhost", true))
	if err != nil {
		panic(err.Error())
	}
	server.Run()
	defer server.httpsListener.Close()
	address := server.httpsListener.Addr
	connection, err := tls.Dial("tcp", address, nil)
	if err != nil {
		panic(err.Error())
	}
	_, err = connection.Write(fuzz)
	if err != nil {
		return 0
	}
	response, err := ioutil.ReadAll(connection)
	if err != nil {
		return 0
	}
	err = connection.Close()
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(response))
	if len(fields) < 2 {
		panic("invalid HTTP response")
	}
	code := fields[1]
	if code == "500" {
		panic("server panicked")
	}
	return 1
}
