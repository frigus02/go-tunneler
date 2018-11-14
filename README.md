# Go Tunneler

> A layer 4 reverse proxy, which forwards packets through an HTTP tunnel

**This is an example implementation and not production ready.**

## Use case

A client is deployed to an environment where it cannot connect to the server directly. It has to use an HTTP proxy. However the client application does not support HTTP proxies.

A layer 7 reverse proxy, which works with the `Host` header, could be a solution. However it would need to terminate the TLS connection in order to read the header. Therefore the client application would have to trust the reverse proxy's certificate.

Another option is a layer 4 proxy. Change the DNS resolution of example.com to resolve to the IP of the Go Tunneler. Using SNI the Go Tunneler will figure out the destination server's hostname. It then opens an HTTP tunnel to a known proxy server and sends all TCP packets through the tunnel.

```
+-------------+
|    Client   |
+------+------+
       |
       | GET https://example.com (DNS: example.com -> go-tunneler)
       v
+------+------+
| Go Tunneler |
+------+------+
       |
       | CONNECT example.com:443 HTTP/1.1 [then tunnel packets]
       v
+------+------+
| HTTP Proxy  |
+------+------+
       |
       | GET https://example.com
       v
+-------------+
|   Server    |
+-------------+
```

## Test this project

1. Setup a local HTTP proxy (e.g. Fiddler)

2. Add the following to your _hosts_ file:

    ```
    127.0.0.1 example.com
    ```

    Make sure the HTTP proxy can still correctly resolve the address. E.g. Fiddler has it's own hosts file option under the _Tools / HOSTS..._ menu.

3. Start the to Go Tunneler:

    ```sh
    go run main.to
    ```

4. Open your browser and go to https://example.com:8443
