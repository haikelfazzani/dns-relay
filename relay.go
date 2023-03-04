package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSRelay struct {
	addrClient   string
	addrResolver string
}

func (dnsRelay *DNSRelay) Start() {

	parser := dnsmessage.Parser{}
	udpAddr, err := net.ResolveUDPAddr("udp", dnsRelay.addrClient)

	if err != nil {
		panic(err)
	}

	packetConn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		panic(err)
	}

	defer packetConn.Close()

	for {
		buff := make([]byte, 512)
		n, clientAddr, err := packetConn.ReadFromUDP(buff)

		if err != nil {
			panic(err)
		}

		go func() {
			parser.Start(buff)
			q, _ := parser.Question()
			fileHostsInfo := dnsRelay.CheckFileHosts(q.Name.String())

			if len(fileHostsInfo) < 1 {
				response := dnsRelay.Forward(buff[:n], clientAddr)
				n, err = packetConn.WriteToUDP(response, clientAddr)
			} else {
				response := dnsRelay.CreateMessage(buff[:n])
				n, err = packetConn.WriteToUDP(response, clientAddr)
			}

			if err != nil {
				panic(err)
			}
		}()
	}
}

func (dnsRelay *DNSRelay) Forward(req []byte, clientAddr *net.UDPAddr) []byte {
	conn, err := net.Dial("udp", dnsRelay.addrResolver)

	if err != nil {
		panic(err)
	}

	_, err = conn.Write(req)

	if err != nil {
		panic(err)
	}

	buff := make([]byte, 512)
	n, err := conn.Read(buff)

	if err != nil {
		panic(err)
	}

	conn.Close()

	return buff[:n]
}

func (dnsRelay *DNSRelay) CheckFileHosts(hostname string) []string {

	cmd := exec.Command("bash", "-c", "grep -Fx '0.0.0.0       "+strings.TrimSuffix(hostname, ".")+"' /etc/hosts")

	stdout, _ := cmd.Output()

	// 0.0.0.0	example.com
	fields := strings.Split(string(stdout), "       ")
	fmt.Printf("\nstdout ===> %v\n \n%v\n", fields, hostname)

	if len(fields) > 1 {
		return fields
	}

	return []string{}
}

func (dnsRelay *DNSRelay) CreateMessage(req []byte) []byte {

	parser := dnsmessage.Parser{}
	header, _ := parser.Start(req)
	questions, _ := parser.AllQuestions()

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			Response:         true,
			Authoritative:    true,
			ID:               header.ID,
			OpCode:           header.OpCode,
			RCode:            header.RCode,
			RecursionDesired: header.RecursionDesired,
		},
		Questions: questions,
		Answers: []dnsmessage.Resource{
			{
				Header: dnsmessage.ResourceHeader{
					Name:  questions[0].Name,
					Type:  dnsmessage.TypeA,
					Class: dnsmessage.ClassINET,
				},
				Body: &dnsmessage.AResource{A: [4]byte(net.IPv4(127, 0, 0, 1))},
			},
		},
	}

	fmt.Printf("\nmsg ===>%v\n%v\n ", header, questions)

	response, _ := msg.Pack()
	return response
}
