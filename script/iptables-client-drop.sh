#!/bin/bash

iptables -A OUTPUT -p tcp --dport 8888 -j DROP