# DuckDNS operator

A kubernetes operator to sync kubernetes ingress Loadbalancer IP addresses to duckdns 

[DuckDNS official website](https://www.duckdns.org/)

## Usage

Deploy using kustomize

Clone this repo

```
git clone https://github.com/kmjayadeep/duckdns-operator.git; cd duckdns-operator
```

Edit the file `kustomize/duckdns.env` and add your domain names and duckdns token there


Apply kustomize on kubernetes


```
kubectl apply -k kustomize/
```
