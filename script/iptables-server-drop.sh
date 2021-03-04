#!/bin/bash

iptables -A INPUT -p tcp --dport 8888 -j DROP