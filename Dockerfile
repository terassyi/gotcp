FROM golang:latest

RUN apt update -y && \
	apt install -y iptables tcpdump


CMD ["/bin/bash"]