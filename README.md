# DNS relay

# Usage
```go
package main

func main() {

	dnsRelay := &DNSRelay{
		raddr:       "8.8.8.8:53",
		laddr:       "127.0.0.1:1053",
		useHosts:    true,
		useWildCard: true,
	}

	println("DNS server started: " + dnsRelay.laddr)

	dnsRelay.Start()
}
```

```shell
dig @127.0.0.1 -p 1053 google.com
```

# Todo

- [ ] cache
- [ ] DNSSec
- [ ] Ratelimit

# License
MIT