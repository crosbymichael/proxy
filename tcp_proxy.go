package proxy

import (
	"fmt"
	"github.com/miekg/dns"
	"io"
	"log"
	"net"
	"sync"
	"syscall"
)

var dnsClient = &dns.Client{}
var server = "172.17.42.1:53"

func tcpProxy(c chan *net.TCPConn, backend *Backend) {
	var (
		group = &sync.WaitGroup{}
		msg   = &dns.Msg{}
	)
	msg.SetQuestion(backend.Query, dns.TypeSRV)

	for conn := range c {
		if err := handleConnection(conn, group, msg); err != nil {
			log.Println(err)
		}
	}
}

func handleConnection(conn *net.TCPConn, group *sync.WaitGroup, msg *dns.Msg) error {
	defer func() {
		conn.CloseRead()
		conn.CloseWrite()
		conn.Close()
	}()

	reply, _, err := dnsClient.Exchange(msg, server)
	if err != nil {
		return err
	}
	if len(reply.Answer) == 0 {
		return fmt.Errorf("no backends avaliable for %s", msg.Question[0])
	}
	first := reply.Answer[0]
	v, ok := first.(*dns.SRV)
	if !ok {
		return fmt.Errorf("dns response not valid SRV record")
	}
	port := v.Port

	extra := reply.Extra[0]
	ev, ok := extra.(*dns.A)
	if !ok {
		return fmt.Errorf("dns extra not valid A record")
	}
	host := ev.A.String()

	destination, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	dest := destination.(*net.TCPConn)
	group.Add(2)

	go transfer(conn, dest, group)
	go transfer(dest, conn, group)

	group.Wait()

	return nil
}

func transfer(from, to *net.TCPConn, group *sync.WaitGroup) {
	defer group.Done()
	if _, err := io.Copy(to, from); err != nil {
		if err, ok := err.(*net.OpError); ok && err.Err == syscall.EPIPE {
			from.CloseWrite()
		}
		// TODO: report errors back
		log.Println(err)
	}
	to.CloseRead()
}

func ProxyConnections(backend *Backend) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: backend.IP, Port: backend.Port})
	if err != nil {
		return err
	}
	defer listener.Close()

	var (
		errorCount  int
		connections = make(chan *net.TCPConn, backend.ConnectionBuffer)
	)

	for i := 0; i < backend.MaxConcurrent; i++ {
		go tcpProxy(connections, backend)
	}

	// start the main event loop for the proxy
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			errorCount++
			log.Println(err)
			continue
		}
		connections <- conn
	}
	return nil
}
