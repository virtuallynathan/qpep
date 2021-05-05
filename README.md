# qpep
Fixing the go version of qpep


# Using Standalone QPEP
 
>:warning: **Disclaimer**: While it is possible to configure and run QPEP outside of the testbed environment, this is discouraged for anything other than experimental testing. The current release of QPEP is a proof-of-concept research tool and, while every effort has been made to make it secure and reliable, it has not been vetted sufficiently for its use in critical satellite communications. Commercial use of this code in its current state would be **exceptionally foolhardy**. When QPEP reaches a more mature state, this disclaimer will be updated.


## Setting Up The Network

The testbed comes with a pre-built and pre-configured QPEP deployment. However, if you wish to use QPEP outside of the test bed this is possible.

You will need at least two machines and ideally three, a QPEP client and a QPEP server are required and a client workstation is optional but recommended. The QPEP client must be able to talk to the QPEP server (e.g. must be able to ping it / initiate UDP connections to open ports on the QPEP server). The client workstation must be configured to route all TCP traffic through the QPEP client.

If you wish to route traffic bi-directionally (e.g. correctly optimize incoming ssh connections to the QPEP client workstation from the internet) you will need to run a QPEP client and a QPEP server on both sides of the connection.

### Client Setup
The client must be configured to route all incoming TCP traffic to the QPEP server. In *nix systems you can do this using iptables. QPEP by default is configured to accept incoming client connections on port 8080
```bash
$ sysctl -w net.ipv4.ip_forward=1
$ iptables -A PREROUTING -t mangle -p tcp -i [network interface to server] -j TPROXY --on-port 8080 --tproxy-mark 1
$ iptables -A PREROUTING -t mangle -p tcp -i [network interface to workstation] -j TPROXY --on-port 8080 --tproxy-mark 1
$ ip rule add fwmark 1 lookup 100
$ ip route add local 0.0.0.0/0 dev lo table 100
```
### Server Setup
No special routing setup is required for the QPEP server. It listens by default on UDP port 4242. If you would like, you can enable ip forwarding which, depending on the underlying network implementation, may allow for fully transparent proxy implementation.
```bash
$ sysctl -w net.ipv4.ip_forward=1
```
## Running QPEP
### Launching the QPEP Client
To run QPEP in client mode once you've set the appropriate IP tables rules:
```bash
$ ./qpep -client -gateway [IP of QPEP server]
```
### Launching the QPEP Server
To run QPEP in server mode
```bash
$ ./qpep
```
### Changing Further QUIC Parameters
QPEP comes with a forked and modified version of the quic-go library which allows for altering some basic constants in the default QUIC implementation. These are provided as command-line flags and can be implemented on both the QPEP server and QPEP client. You can use ```qpep -h``` to see basic help output. The available options are:
* ```-acks [int]``` Sets the number of ack-eliciting packets per ack. The default ratio is 10:1.
* ```-congestion [int]``` Sets the size of the initial QUIC congestion window in number of QUIC packets. Defaults to 4.
* ```-multistream [bool]``` Enables multiplexing QUIC streams inside a meta-session. Default is true.
* ```-ackDelay [int]``` Maximum number of miliseconds to hold back an ack for decimation. Default is 25.
* ```-varAckDelay [float]``` Variable number of miliseconds to try and hold back an ack for decimation, as multiple of RTT. Default is 0.25.
* ```-minBeforeDecimation [int]``` Minimum number of packets sent before initiating any ack decimation. Default is 100.
* ```-client [bool]``` runs QPEP in client mode. Default is false.
* ```-gateway [ip]``` sets the gateway address for a QPEP client to connect to. Default is 192.18.0.254 but you will probably need to set it yourself based on your network config.
