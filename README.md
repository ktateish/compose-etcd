# compose-etcd: Generates docker-compose.yaml for etcd cluster w/ TLS support

## Requirement

* The Go Language
* GNU Make

## Step1: Prepare Configuration

```bash
$ git clone https://github.com/ktateish/compose-etcd.git
$ cd compose-etcd
$ cp config.yaml.example config.yaml
$ vi config.yaml
# Edit to suit your needs
```

* config.yaml example
```yaml
template:
  domain: example.org
  token: mysecrettoken
spec:
- name: etcd0
- name: etcd1
- name: etcd2
```

## Step2: Generate docker-compose.yaml

```bash
$ make
go build gen.go
./gen < config.yaml
$ ls -1 compose/*/*
compose/etcd0/docker-compose.yaml
compose/etcd1/docker-compose.yaml
compose/etcd2/docker-compose.yaml
```

## Step3: Prepare your certs

* You need following to run etcd with https:
  * Self signed CA public key (ca.pem)
  * Etcd server node's private key (etcd-key.pem)
  * Its public key signed by the CA (etcd.pem)
* Generated docker-compose.yaml expects that ca.pem, etcd-key.pem, etcd.pem 
  are in certs/ directory (e.g. etcd0/certs/ca.pm)
* Your key pair for etcd nodes can be a single key pair, but they must have
  all hostnames/IP addresses that you will access to.
  * e.g. when you have three nodes, 'etcd0', 'etcd1', 'etcd2' and they have
    the same DNS name, 'etcd', pointing them like DNS load balancing.
    Your cert should have 'etcd', 'etcd0', 'etcd1', 'etcd2' in its extension
    field.
    And if you wish to access them from localhost, you have to add localhost.
* Further more, you need other certs for your client.
* Generating all these keys are tiring but 'ezcerts' will help you
* copy ca.pem, etcd.pem, etcd-key.pem to certs/ directory in each etcdN/
```bash
$ mkdir compose/etcd0/certs
$ cp /path/to/ca.pem compose/etcd0/certs
$ cp /path/to/etcd.pem compose/etcd0/certs
$ cp /path/to/etcd-key.pem compose/etcd0/certs
```

## Step4: copy each directory to server node

```bash
$ scp -r compose/etcd0 etcd0:/var/lib/etcd
$ scp -r compose/etcd1 etcd1:/var/lib/etcd
$ scp -r compose/etcd2 etcd2:/var/lib/etcd
```

## Step5: docker-compose up on each node

```bash
etcd0$ cd /var/lib/etcd/etcd0
etcd0$ docker-compose up -d && docker-compose logs -f
Attaching to etcd0_etcd0_1
etcd0_1  | 2016-08-28 14:29:15.912599 I | flags: recognized and used environment variable ETCD_ADVERTISE_CLIENT_URLS=https://etcd0:2379
etcd0_1  | 2016-08-28 14:29:15.912763 I | flags: recognized and used environment variable ETCD_CERT_FILE=/certs/etcd.pem
etcd0_1  | 2016-08-28 14:29:15.912773 I | flags: recognized and used environment variable ETCD_CLIENT_CERT_AUTH=true
etcd0_1  | 2016-08-28 14:29:15.912786 I | flags: recognized and used environment variable ETCD_DATA_DIR=/data/etcd1.etcd

...

```
