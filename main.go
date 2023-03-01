package main

import "fmt"

func main() {

	server := &dnsServer{
		forwardDnsResolver: "8.8.8.8:53", // the address of the DNS server to forward requests to
	}

	// start the DNS server
	err := server.Start("127.0.0.1:1053")
	if err != nil {
		fmt.Printf("Failed to start DNS server: %s\n", err)
		return
	}
}
