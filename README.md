# qpep
Working on improving the Go standalone implementation of qpep, improving documentation. Original full repo: https://github.com/ssloxford/qpep


# Using Standalone QPEP
 
>:warning: **Disclaimer**: While it is possible to configure and run QPEP outside of the testbed environment, this is discouraged for anything other than experimental testing. The current release of QPEP is a proof-of-concept research tool and, while every effort has been made to make it secure and reliable, it has not been vetted sufficiently for its use in critical satellite communications. Commercial use of this code in its current state would be **exceptionally foolhardy**. When QPEP reaches a more mature state, this disclaimer will be updated.


## Setting Up The Network

The testbed comes with a pre-built and pre-configured QPEP deployment. However, if you wish to use QPEP outside of the test bed this is possible.

You will need at least two machines and ideally three, a QPEP client and a QPEP server are required and a client workstation is optional but recommended. The QPEP client must be able to talk to the QPEP server (e.g. must be able to ping it / initiate UDP connections to open ports on the QPEP server). The client workstation must be configured to route all TCP traffic through the QPEP client.

If you wish to route traffic bi-directionally (e.g. correctly optimize incoming ssh connections to the QPEP client workstation from the internet) you will need to run a QPEP client and a QPEP server on both sides of the connection.

### Client Setup
The client must be configured to route all incoming TCP traffic to the QPEP server. In *nix systems you can do this using nftables. QPEP by default is configured to accept incoming client connections on port 8080.
```bash
$ sysctl -w net.ipv4.ip_forward=1
$ sysctl -w net.core.rmem_max=2500000
$ nftables -f nftables.conf
$ ip rule add fwmark 0x233 lookup 100
$ ip route add local 0.0.0.0/0 dev lo table 100
```

A systemd service script is included with helpful start/stop/reload options. IPs/prefixes may be excluded from proxying by editing the list in nftables.conf. 

### Server Setup
No special routing setup is required for the QPEP server. It listens by default on UDP port 4242. If you would like, you can enable ip forwarding which, depending on the underlying network implementation, may allow for fully transparent proxy implementation.
```bash
$ sysctl -w net.ipv4.ip_forward=1
```
## Running QPEP
### Launching the QPEP Client
To run QPEP in client mode once you've set the appropriate IP tables rules:
```bash
$ sysctl -w net.core.rmem_max=2500000
$ ./qpep -client -gateway [IP of QPEP server]
```
### Launching the QPEP Server
To run QPEP in server mode
```bash
$ sysctl -w net.core.rmem_max=2500000
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


## References in Publications 
QPEP and the corresponding testbed were both designed to encourage academic research into secure and performant satellite communications. We would be thrilled to learn about projects you're working on academically or in industry which build on QPEP's contribution!

If you use QPEP, the dockerized testbed, or something based on it, please cite the conference paper which introduces QPEP:
> Pavur, James, Martin Strohmeier, Vincent Lenders, and Ivan Martinovic. QPEP: An Actionable Approach to Secure and Performant Broadband From Geostationary Orbit. Network and Distributed System Security Symposium (NDSS 2021), February 2021. [https://ora.ox.ac.uk/objects/uuid:e88a351a-1036-445f-b79d-3d953fc32804](https://ora.ox.ac.uk/objects/uuid:e88a351a-1036-445f-b79d-3d953fc32804).

## License
The Clear BSD License

Copyright (c) 2020 James Pavur.

All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted (subject to the limitations in the disclaimer
below) provided that the following conditions are met:

     * Redistributions of source code must retain the above copyright notice,
     this list of conditions and the following disclaimer.

     * Redistributions in binary form must reproduce the above copyright
     notice, this list of conditions and the following disclaimer in the
     documentation and/or other materials provided with the distribution.

     * Neither the name of the copyright holder nor the names of its
     contributors may be used to endorse or promote products derived from this
     software without specific prior written permission.

NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY
THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND
CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR
BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER
IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

“Commons Clause” License Condition v1.0

The Software is provided to you by the Licensor under the License, as defined below, subject to the following condition.

Without limiting other conditions in the License, the grant of rights under the License will not include, and the License does not grant to you, the right to Sell the Software.

For purposes of the foregoing, “Sell” means practicing any or all of the rights granted to you under the License to provide to third parties, for a fee or other consideration (including without limitation fees for hosting or consulting/ support services related to the Software), a product or service whose value derives, entirely or substantially, from the functionality of the Software. Any license notice or attribution required by the License must also include this Commons Clause License Condition notice.

## Acknowledgments
[OpenSAND](https://opensand.org/content/home.php) and the [Net4Sat](https://www.net4sat.org/content/home.php) project have been instrumental in making it possible to develop realistic networking simulations for satellite systems.

This project would not have been possible without the incredible libraries developed by the Go community. These libraries are linked as submodules in this git repository. We're especially grateful to the [quic-go](https://github.com/lucas-clemente/quic-go) project.
