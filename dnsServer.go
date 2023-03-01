package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type dnsServer struct {
	forwardDnsResolver string
}

func (s *dnsServer) Start(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	buf := make([]byte, 512)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)

		if err != nil {
			fmt.Printf("Error reading from UDP: %s\n", err)
			continue
		}

		go func() {
			req := buf[:n]

			resp, err := s.processRequest(req)
			if err != nil {
				fmt.Printf("Error processing DNS request: %s\n", err)
				return
			}

			_, err = conn.WriteToUDP(resp, clientAddr)
			if err != nil {
				fmt.Printf("Error writing DNS response: %s\n", err)
				return
			}
		}()
	}
}

func (s *dnsServer) processRequest(req []byte) ([]byte, error) {
	// Check /etc/hosts
	requestMsg := &dns.Msg{}
	err := requestMsg.Unpack(req)
	hostname := strings.TrimSuffix(requestMsg.Question[0].Name, ".")

	println("hostname ===> ", hostname)

	if err != nil {
		log.Fatal(err)
	}

	if len(findHost(hostname)) > 2 {
		resp, err := s.createResponse(requestMsg)
		return resp, err
	}

	// Forward the request to the resolver
	resp, err := s.forward(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *dnsServer) forward(req []byte) ([]byte, error) {
	// create a UDP connection to the forward address
	conn, err := net.Dial("udp", s.forwardDnsResolver)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func (s *dnsServer) createResponse(req *dns.Msg) ([]byte, error) {

	ipType := req.Question[0].Qtype

	// Create a new response message
	resp := new(dns.Msg)

	resp.Id = req.Id

	resp.Rcode = dns.RcodeSuccess
	resp.RecursionAvailable = true
	resp.Authoritative = true

	// Copy the questions from the request to the response
	resp.Question = append(resp.Question, req.Question...)

	// Add a new answer record to the response
	if ipType == 28 {
		rr := &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   req.Question[0].Name,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    300,
			},
			AAAA: net.IPv4zero,
		}

		resp.Answer = append(resp.Answer, rr)
		return resp.Pack()
	}

	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   req.Question[0].Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		A: net.IPv6zero,
	}

	resp.Answer = append(resp.Answer, rr)
	return resp.Pack()
}
