# Gotcp
Gotcp is user space tcp/ip protocol stack implementation by golang for learing purpose.

## About
gotcp gets raw data from network interface with linux pf_packet.
Supporting protocols is brow.
- Ethernet
- ARP
- IPv4
- ICMP
- TCP

Now ping, tcpclient and tcpserver almost work.
But tcpclient and tcpserver don't work between tcp server and client working on os stack. 
I'm not sure about this problem.

## Run
Gotcp run only on linux. If you run on virtual machines, `Vagrantfile` is in `/vms` folder.
run `go build` to build.
After building finished, run `./gotcp [command]`.
### Usage
```
$ ./gotcp
Usage: gotcp <flags> <subcommand> <subcommand args>

Subcommands:
	commands         list all command names
	dump             dump
	flags            describe all known top-level flags
	help             describe subcommands and their syntax
	ping             ping
	tcpclient        tcp client
	tcpserver        tcp server
```

## Future
Now gotcp can't communicate with tcp server or client program working on OS stack, I want to solve this problem.
