# Gotcp
Gotcp is user space tcp/ip protocol stack implementation by golang for learing purpose.
This project is inspired by [microps](https://github.com/pandax381/microps)

## Features
Gotcp work on Linux only. To run this, You have to be root.
Supported protocol is berow.
- Ethernet
	- tuntap
	- ap_packet
- ARP
- IPv4
- ICMP
- TCP

## Tutorial
You can run Gotcp with virtual machines or docker containers.

### Environment
#### Virtual machine
I prepare the virtual machines in `vm/`. You can setup vm via Vagrant.

#### Container
I also prepare `docker-compose.yml`. This way is easier than using virtual machines.

### Run
To buil this, execute `go build .`

I prepare some command to run sample application.
```shell
./gotcp help
Usage: gotcp <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        dump             dump
        flags            describe all known top-level flags
        help             describe subcommands and their syntax
        ids              ids
        ping             ping
        tcpclient        tcp client
        tcpserver        tcp server
```

#### ping
```shell
# ./gotcp ping -h
goctp ping -i <interface name> -dest <destination address>:
        send icmp echo request packets and receive reply packets  -debug
        output debug messages
  -dest string
        destination address
  -i string
        interface
```
```shell
# ./gotcp ping -i eth0 -dest 172.20.0.3
172.20.0.3
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=6953) ms
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=1011073) ms
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=2013006) ms
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=3018657) ms
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=4019847) ms
47 bytes from 172.20.0.3: icmp_seq=0 ttl=64 time=%!f(int64=5020866) ms

```

#### tcp client
You can run a sample tcp client application.
```shell
# ./gotcp tcpclient -h
gotcp tcpclient -i <interface name> -addr <ip address> -port <port>
        tcp client to destination host  -addr string
        destination host address
  -debug
        output debug message
  -i string
        interface name
  -port int
        destination host port
```
Before running this application, you have to execute these commands on the client and server side.

- server
	```shell
	$ ethtool -K eth0 tso off gso off # stop kernel nic offloading
	$ cd app/standard/tcp-server-big
	$ go run main.go # run server program
	```
- client
	```shell
	$ mkdir data
	$ head -c 20000 /dev/urandom > data/random-data # generate random data
	$ iptables -A OUTPUT -p tcp --dport 8888 -j DROP # stop processing packets by kernel
	$ ./gotcp tcpclient -debug -i eth0 -addr 172.20.0.3 -port 8888 # run gotcp tcpclient
	```
log
```
INFO[0000] tcp server running at 8888                    command="tcp client"
INFO[0000] [start to listen]                             protocol=tcp
DEBU[0000] [LISTEN]                                      protocol=tcp
DEBU[0003] [SYN_RECVD]                                   protocol=tcp
DEBU[0003] [ESTABLISHED]                                 protocol=tcp
DEBU[0003] [retransmission routine start ]               protocol=tcp
Server> Connection from 172.20.0.3:8888
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 208 bytes
Server> recv all buf 21720 bytes
Server> Write 20272 bytes
DEBU[0008] [CLOSE_WAIT]                                  protocol=tcp
DEBU[0008] [LAST_ACK]                                    protocol=tcp
DEBU[0008] [CLOSED]                                      protocol=tcp
INFO[0008] [received packet is not handled. invalid peer.]  protocol=tcp
Server> close
```

#### tcp server
You can run a sample tcp server application.
```shell
$ ./gotcp tcpserver -h
gotcp tcpserver -i <interface name> -port <port>
        tcp server binding port  -debug
        output debug message
  -i string
        interface
  -port int
        binding port
```
Befor you run this application, you have to execute these commands.
- server
	```shell
	$ iptables -A INPUT -p tcp --dport 4000:65000 -j DROP # stop packets processing by kernel
	$ ./gotcp tcpserver -debug -i eth0 -port 8888
	```
- client
	```shell
	$ ethtool -K eth0 tso off gso off # stop kernel nic offloading
	$ cd app/standard/tcp-client-big
	$ go run main.go # run client program
	```

log
```
INFO[0000] tcp server running at 8888                    command="tcp client"
INFO[0000] [start to listen]                             protocol=tcp
DEBU[0000] [LISTEN]                                      protocol=tcp
DEBU[0004] [SYN_RECVD]                                   protocol=tcp
DEBU[0005] [ESTABLISHED]                                 protocol=tcp
DEBU[0005] [retransmission routine start ]               protocol=tcp
Server> Connection from 172.20.0.3:8888
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 1448 bytes
Server> Read 208 bytes
Server> recv all buf 21720 bytes
Server> Write 20272 bytes
DEBU[0015] [CLOSED]                                      protocol=tcp
Server> close
```

## License
Gotcp is under the MIT License: See [LICENSE](./LICENSE) file.
