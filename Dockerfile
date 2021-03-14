FROM golang:latest

RUN apt update -y && \
	apt install -y iptables tcpdump ethtool

#RUN ethtool -K eth0 tso off gso off


CMD ["/bin/bash"]