# p2p_proxy
This the git repo for independent project, a p2p proxy network. Its goal is to protect users' privacy and hide
their online identity by forwarding their network requests in a random manner.
# Requirements
- Linux machine
- Go version go1.8.1 (If want to build the project)
# How to start using p2p_proxy
0. Run following command to get the executable
```
go build peerproxy.go
```
1. change 'peerproxy' to executable if necessary (for linux, 'chmod +x peerproxy')
2. Configure your local computer proxy to 'localhost:3129' (for macOS, system preference -> network-> advanced -> proxies-> web proxy, type 'localhost' in Web Proxy Server, and '3129' as the port)
3. Open your terminal, run peerproxy with one of the available peers. For example, type
```
/.peerproxy [available peer IP]
```
