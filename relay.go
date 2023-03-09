package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSRelay struct {
	laddr       string
	raddr       string
	useHosts    bool
	useWildCard bool
}

func (dnsRelay *DNSRelay) Start() {

	parser := dnsmessage.Parser{}
	udpAddr, err := net.ResolveUDPAddr("udp", dnsRelay.laddr)

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
		n, laddr, err := packetConn.ReadFromUDP(buff)

		if err != nil {
			panic(err)
		}

		go func() {
			parser.Start(buff)
			q, _ := parser.Question()

			foundedIp := ""

			if dnsRelay.useHosts {
				foundedIp = dnsRelay.ResolveFromHostsFile("/etc/hosts", q.Name.String())
				if foundedIp != "" {
					response := dnsRelay.NewMessage(buff[:n], foundedIp)
					n, err = packetConn.WriteToUDP(response, laddr)
					return
				}
			}

			response := dnsRelay.Resolve(buff[:n], laddr)
			n, err = packetConn.WriteToUDP(response, laddr)

			if err != nil {
				panic(err)
			}
		}()
	}
}

func (dnsRelay *DNSRelay) Resolve(req []byte, laddr *net.UDPAddr) []byte {
	conn, err := net.Dial("udp", dnsRelay.raddr)

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

func (dnsRelay *DNSRelay) ResolveFromHostsFile(filePath string, hostname string) string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "#") && line != "" {
			r := regexp.MustCompile(`[^\s"]+`)
			fields := r.FindAllString(line, -1)

			if dnsRelay.useWildCard && fields[1] == strings.Trim(hostname, ".") {
				return fields[0]
			}

			if strings.HasPrefix(fields[1], strings.Trim(hostname, ".")) {
				return fields[0]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	return ""
}

func (dnsRelay *DNSRelay) NewMessage(req []byte, ip string) []byte {

	ipbytes := [4]byte(net.IPv4(0, 0, 0, 0))

	if ip != "" {
		ipbytes = [4]byte(net.IPv4(ip[0], ip[1], ip[2], ip[3]))
	}

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
				Body: &dnsmessage.AResource{A: ipbytes},
			},
		},
	}

	response, _ := msg.Pack()
	return response
}
