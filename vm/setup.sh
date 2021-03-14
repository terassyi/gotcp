#!/bin/bash

# update
sudo apt update -y
sudo apt install ethtool

# install golang
wget https://golang.org/dl/go1.15.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.15.7.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

sudo ethtool -K eth0 tso off gso off