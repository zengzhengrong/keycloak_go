package main

import (
	"context"
	"flag"
	"keycloak-example/keycloak"
	"log"
	"net"
	"reflect"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	port = ":8000"
)

var (
	apiClientID     string   = "example-rpc-go"
	apiClientSecret string   = "050aead4-85d0-4bb4-b099-ee3184cc42bc"
	publicRole      []string = []string{"user", "admin"}
	securedRole     []string = []string{"admin"}
	serverTLS                = flag.Bool("server_tls", false, "Connection uses TLS if true, else plain TCP")
)

type server struct {
	keycloak.UnimplementedKeycloakServiceServer
}

func (s *server) Public(ctx context.Context, in *empty.Empty) (*keycloak.Reply, error) {
	log.Printf("Received: Public")
	ms, err := authAPI(ctx, publicRole)
	if err != nil {
		return nil, err
	}
	if ms != "" {
		var statuscode int32
		statuscode = 403
		if strings.Contains(ms, "401") {
			statuscode = 401
		}
		return &keycloak.Reply{Message: ms, Statuscode: statuscode}, nil
	}

	return &keycloak.Reply{Message: "I am Public", Statuscode: 200}, nil
}

func (s *server) Secured(ctx context.Context, in *empty.Empty) (*keycloak.Reply, error) {
	log.Printf("Received: Secured")
	ms, err := authAPI(ctx, securedRole)
	if err != nil {
		return nil, err
	}
	if ms != "" {
		var statuscode int32
		statuscode = 403
		if strings.Contains(ms, "401") {
			statuscode = 401
		}
		return &keycloak.Reply{Message: ms, Statuscode: statuscode}, nil
	}
	return &keycloak.Reply{Message: "I am Secured", Statuscode: 200}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	// register interceptor
	opts = append(opts, grpc.UnaryInterceptor(interceptor))
	if *serverTLS {
		// add server tls
		creds, err := credentials.NewServerTLSFromFile("./keys/server.crt", "./keys/server.key")
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	keycloak.RegisterKeycloakServiceServer(s, &server{})
	log.Printf("Start Server %v ,TLS: %v", port, *serverTLS)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func rolecheck(id, secret *string, rs []string, token oauth2.Token) (bool, string) {
	result := keycloak.Client(id, secret, rs, token)
	resultValue := reflect.ValueOf(result).Elem()
	pass := resultValue.FieldByName("pass").Bool()
	messgae := resultValue.FieldByName("messgae").String()
	return pass, messgae
}

// authAPI , Verify API by Token
func authAPI(ctx context.Context, rs []string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", grpc.Errorf(codes.Unauthenticated, "Context error")
	}
	// get token
	value, ok := md["authorization"]
	token := oauth2.Token{
		// Split token Bear
		AccessToken: strings.Split(value[0], " ")[1],
	}
	_, errms := rolecheck(&apiClientID, &apiClientSecret, rs, token)

	return errms, nil
}

// auth , Verify Token

// interceptor , TODO: The permission code would be here
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// err := auth(ctx)

	return handler(ctx, req)
}
