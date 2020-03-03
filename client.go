package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"keycloak-example/keycloak"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/grpclog"
)

const (
	address = ":8000"
)

var (
	clientTLS = flag.Bool("client_tls", false, "Connection uses TLS if true, else plain TCP")
)

type customAuth struct{}

// NewClientTLSFromFile , client tls
func NewClientTLSFromFile(certFile, serverNameOverride string) (credentials.TransportCredentials, error) {
	b, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}
	return credentials.NewTLS(&tls.Config{ServerName: serverNameOverride, InsecureSkipVerify: true, RootCAs: cp}), nil
}

func (c customAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	keycloaktoken := keycloak.LoginKeycloak()
	token := oauth2.Token{
		AccessToken:  keycloaktoken.AccessToken,
		RefreshToken: keycloaktoken.RefreshToken,
		TokenType:    keycloaktoken.TokenType,
	}
	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

func (c customAuth) RequireTransportSecurity() bool {

	return false
}
func main() {
	flag.Parse()
	// Set up a connection to the server.
	var opts []grpc.DialOption
	var creds credentials.PerRPCCredentials
	log.Printf("Client TLS : %v", *clientTLS)
	if *clientTLS {
		// add tls
		credstls, err := NewClientTLSFromFile("./keys/server.crt", "")
		if err != nil {
			grpclog.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credstls))
		// Login to get token
		token := keycloak.LoginKeycloak()
		oauthtoken := oauth2.Token{
			AccessToken:  token.AccessToken,
			TokenType:    token.TokenType,
			RefreshToken: token.RefreshToken,
		}
		// NewOathAccess must be required RequireTransportSecurity set true
		creds = oauth.NewOauthAccess(&oauthtoken)
	} else {
		// Insecure
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithBlock())
		creds = customAuth{}
	}

	// custom auth : input token to ctx
	opts = append(opts, grpc.WithPerRPCCredentials(creds))

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := keycloak.NewKeycloakServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	// log.Printf("%v", )
	r, err := c.Public(ctx, &empty.Empty{})
	log.Printf("Keycloak: status<%v>, messgae<%s>", r.GetStatuscode(), r.GetMessage())
	rs, err := c.Secured(ctx, &empty.Empty{})
	if err != nil {
		log.Fatalf("could not keycloak: %v", err)
	}

	log.Printf("Keycloak: status<%v>, messgae<%s>", rs.GetStatuscode(), rs.GetMessage())

}
