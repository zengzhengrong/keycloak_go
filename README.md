## Secured go-grpc api with keycloak 

This repo is Explore solution & Testing 

## First setting your keycloak

How to insall keycloak ?

containerizd keycloak see below

- [keycloak](https://github.com/keycloak/keycloak-containers)

Clone and import Realm ``` realm-export.json ``` to keycloak manage Page 

Make sure user ```zzr``` have client roles view-users , view-clients in realm-management  
The gocloak library will use these roles

## Second setting keycloak config in  keycloak/keycloak_client.go

```
const (
	hostname     string = "http://keycloak-ingress-l7-http.keycloak.192.168.1.203.xip.io"
	clientid     string = "flutter_dev"
	clientSecret string = ""
	realm        string = "demo"
	username     string = "zzr"
	password     string = "pass"
)
```
Setting your hostname of kaycloak host ,Others is by default 


## Run 

1. Insecure transport
```
go run server.go
// 2020/03/03 07:12:26 Start Server :8000 ,TLS: false
```

```
go run client.go
// 2020/03/03 07:12:51 Client TLS : false
// 2020/03/03 07:12:51 Keycloak: status<200>, messgae<I am Public>
// 2020/03/03 07:12:51 Keycloak: status<200>, messgae<I am Secured>
...

```
2. With TLS/SSL tranport
Use server_tls & client_tls args
```
go run server.go -server_tls=true
// 2020/03/03 07:14:47 Start Server :8000 ,TLS: true
```
```
go run client.go -client_tls=true
// 2020/03/03 07:16:05 Client TLS : true
// 2020/03/03 07:16:05 Keycloak: status<200>, messgae<I am Public>
// 2020/03/03 07:16:05 Keycloak: status<200>, messgae<I am Secured>
```

## Other Setting 

default address port ":8000"  

server.go  

```
var (
	apiClientID     string   = "example-rpc-go"
	apiClientSecret string   = "050aead4-85d0-4bb4-b099-ee3184cc42bc"
	publicRole      []string = []string{"user", "admin"}
	securedRole     []string = []string{"admin"}
```

apiClientID & apiClientSecret must required , When token access api(grpc) it would check out the token with apiClient , Whether or not  have roles in this client


## With flutter 

More information in [flutter_grpc](https://github.com/zengzhengrong/flutter_grpc)