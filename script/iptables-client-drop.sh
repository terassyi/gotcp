#!/bin/bash

iptables -A OUTPUT -p tcp --dport 60000:65535 -j DROP