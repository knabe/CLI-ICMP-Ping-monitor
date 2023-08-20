# CLI-ICMP-Ping-monitor

A simple golang app for monitoring a series of IPs 

## Getting started

Build the app

```
make
```

Run the app

```
./ping-monitor 1s iplist.txt
```

ip table example
iplist.txt
```
192.168.1.1,router
192.168.1.2, example1
192.168.1.3, example2
```