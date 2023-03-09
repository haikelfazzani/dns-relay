package main

func main() {

	dnsRelay := &DNSRelay{
		raddr:    "8.8.8.8:53",
		laddr:    "127.0.0.1:1053",
		useHosts: true,
	}

	println("DNS server started: " + dnsRelay.laddr)

	dnsRelay.Start()
}
