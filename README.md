# ToastCloud

Toastcloud is a self - hosted platform to run and auto-scale serverless code instances. An instance is started by the first request and can then be joined by multiple other requests. HTTP, Websocket and SSH requests are supported. Joining the same running code instance with different kind of requests is supported. You can set the timeout and maximum number of joiners for each instance in order to autoscale them naturally. You can also make a request force join a particular instance with its ID. You can setup your own OS image to support any programming language.

# Motivation

At Toastate, we need a way to quickly deploy and run autoscaled code instances. We need those instance to be reachable by multiple request of distinct types. We could not vendor lock clients from our web agency. This is why we developed Toastcloud> We believe it can also be useful for other projects and this is why we release it to the community. We also provide a hosted version at toastcloud.toastate.com.

# Setup

## Standalone

## Multi - Nodes and Multi - Cloud

### IP Addresses

To maintain security, IP addresses of nodes must be in a private CIDR (10.0.0.0/8 or 172.16.0.0/12 or 192.168.0.0/16). This means they must be in the same private network or VPN. Toatscloud will throw an error if it is not the case.

# Usage

# Full Configuration example