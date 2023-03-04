package main

func main() {

	dnsRelay := &DNSRelay{
		addrResolver: "8.8.8.8:53",
		addrClient:   "127.0.0.1:1053",
	}

	println("DNS server started" + dnsRelay.addrClient)

	dnsRelay.Start()
}
