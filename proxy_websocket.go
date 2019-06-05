package mqtt

import (
	"bytes"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ProxyWS struct {
	Ws          *websocket.Conn
	MessageType int
	buf         *bytes.Buffer
}

// 
func NewProxyWS(uri *url.URL, headers http.Header) (ProxyWS, error) {
	d := websocket.Dialer{}

	myproxy := getProxy(uri.Host, uri.Scheme)
	if myproxy != nil {
		myproxyUrl := &url.URL{
			Scheme: myproxy.Scheme,
			Host:   myproxy.Host,
			Path:   myproxy.Path,
		}
		d.Proxy = http.ProxyURL(myproxyUrl)
	}

	headers.Add("Origin", uri.Host)
	headers.Add("Sec-WebSocket-Protocol", "mqtt")

	dl, _, err := d.Dial(uri.String(), headers)

	proxyWs := ProxyWS{
		Ws:          dl,
		MessageType: websocket.BinaryMessage,
		buf:         new(bytes.Buffer),
	}
	return proxyWs, err
}

func getProxy(addr string, scheme string) *url.URL {
	if scheme == "wss" {
		httpsProxy := os.Getenv("HTTPS_PROXY")
		if httpsProxy == "" {
			httpsProxyUrl := getProxy(addr, "ws")
			httpsProxy = httpsProxyUrl.Host
		}
	}

	proxy := os.Getenv("HTTP_PROXY")
	if proxy == "" {
		proxy = os.Getenv("http_proxy")

	}
	// test for no proxy, takes a comma separated list with substrings to match
	if proxy != "" {
		noproxy := os.Getenv("NO_PROXY")
		if noproxy == "" {
			noproxy = os.Getenv("no_proxy")
		}
		if noproxy != "" {
			nplist := strings.Split(noproxy, ",")
			for _, s := range nplist {
				if containsIgnoreCase(addr, s) {
					proxy = ""
					break
				}
			}
		}
	}

	if proxy != "" {
		url, err := url.Parse(proxy)
		if err == nil {
			addr = url.Host
		}
	}

	if proxy == "" {
		return nil
	}

	pr, err := url.Parse(proxy)
	if err != nil {
		WARN.Printf("error parsing proxy %v", err)
	}
	return pr
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (s ProxyWS) Read(b []byte) (n int, err error) {
	if s.buf.Len() > 0 {
		return s.buf.Read(b)
	}

	if s.Ws == nil {
		return 0, nil
	}

	_, r, err := s.Ws.ReadMessage()

	if err != nil {
		return 0, err
	}

	s.buf.Write(r)

	return s.buf.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (s ProxyWS) Write(b []byte) (int, error) {
	if s.Ws == nil {
		return 0, nil
	}

	err := s.Ws.WriteMessage(s.MessageType, b)

	return len(b), err
}

func (s ProxyWS) Close() error {
	err := s.Ws.Close()
	s.Ws = nil

	return err
}

// LocalAddr returns the local network address.
func (s ProxyWS) LocalAddr() net.Addr {
	return s.Ws.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (s ProxyWS) RemoteAddr() net.Addr {
	return s.Ws.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (s ProxyWS) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (s ProxyWS) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (s ProxyWS) SetWriteDeadline(t time.Time) error {
	return nil
}
