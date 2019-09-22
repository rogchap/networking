# Implementing a DNS Client

```bash
usage: go run . [@dns-server] [q-type] host
where:  dns-server  is the IPv4 address of a DNS server [default: 1.1.1.1]  
        q-type      is one of (A, NS, MX, TXT) [default: A]
```

### Example

```bash
$ go run . rogchap.com
;; ANSWER SECTION:
rogchap.com.	300	IN	A	185.199.111.153
rogchap.com.	300	IN	A	185.199.108.153
rogchap.com.	300	IN	A	185.199.109.153
rogchap.com.	300	IN	A	185.199.110.153
```
