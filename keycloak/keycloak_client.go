package keycloak

import (
	"fmt"
	"log"

	"github.com/Nerzal/gocloak/v5"
	"golang.org/x/oauth2"
)

const (
	hostname     string = "http://keycloak-ingress-l7-http.keycloak.192.168.1.203.xip.io"
	clientid     string = "flutter_dev"
	clientSecret string = ""
	realm        string = "demo"
	username     string = "zzr"
	password     string = "pass"
)

// Message is return Keycloak Reusult
type Message struct {
	pass       bool
	statusCode int
	messgae    string
}

// LoginKeycloak , return token
func LoginKeycloak() *gocloak.JWT {
	client := gocloak.NewClient(hostname)
	token, err := client.Login(clientid, clientSecret, realm, username, password)
	if err != nil {
		panic("Login failed:" + err.Error())
	}
	return token
}

// Client , Keycloak client to inspest user permissins
func Client(id, secret *string, rs []string, token oauth2.Token) *Message {
	result := Message{}
	client := gocloak.NewClient(hostname)
	clients, err := client.GetClients(
		token.AccessToken,
		realm,
		gocloak.GetClientsParams{
			ClientID: id,
		})
	if err != nil {
		apiError := err.(*gocloak.APIError)
		log.Printf("%v", apiError)
		result.pass = false
		result.statusCode = apiError.Code
		result.messgae = "GetClients failed:" + apiError.Message
		return &result
	}
	rptResult, err := client.RetrospectToken(token.AccessToken, *clients[0].ClientID, *secret, realm)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		log.Printf("%v", apiError)
		result.pass = false
		result.statusCode = apiError.Code
		result.messgae = "Inspection failed:" + apiError.Message
		return &result
	}
	active := rptResult.Active
	if !*active {
		result.pass = false
		result.statusCode = 403
		result.messgae = "Token is not active"
		// panic("Token is not active")
	}
	// permissions := rptResult.Permissions

	userinfo, err := client.GetUserInfo(token.AccessToken, realm)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		log.Printf("%v", apiError)
		result.pass = false
		result.statusCode = apiError.Code
		result.messgae = "GetUserInfo failed:" + apiError.Message
		return &result
		// panic("GetUserInfo failed:" + err.Error())
	}

	rolemap, err := client.GetRoleMappingByUserID(token.AccessToken, realm, *userinfo.Sub)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		log.Printf("%v", apiError)
		result.pass = false
		result.statusCode = apiError.Code
		result.messgae = "GetRoleMappingByUserID failed:" + apiError.Message
		return &result
		// panic("GetRoleMappingByUserID failed:" + err.Error())
	}
	// vaild in clientmapping
	_, ok := rolemap.ClientMappings[*id]
	if !ok {
		log.Printf("Your dont have any roles in this client :%v", *id)
		result.pass = false
		result.statusCode = 403
		result.messgae = "Your dont have any roles in this client"
		return &result
	}
	clientroles, err := client.GetClientRolesByUserID(token.AccessToken, realm, *rolemap.ClientMappings[*id].ID, *userinfo.Sub)
	if err != nil {
		apiError := err.(*gocloak.APIError)
		log.Printf("%v", apiError)
		result.pass = false
		result.statusCode = apiError.Code
		result.messgae = "GetClientRolesByUserID failed:" + apiError.Message
		return &result
		// panic("GetClientRolesByUserID failed:" + err.Error())
	}
	// Inspection role in clientroles
	var haverole = false

	for _, r := range clientroles {

		for _, rr := range rs {
			if *r.Name != rr {
				continue
			}
			haverole = true
			break
		}

		if haverole {
			break
		}
	}
	if !haverole {
		log.Printf("Yor don't have %v of role of %v in %v ", rs, *id, realm)
		result.pass = false
		result.statusCode = 403
		result.messgae = fmt.Sprintf("Yor don't have %v of role of %v in %v ", rs, *id, realm)
		return &result
	}

	result.pass = true
	result.statusCode = 200
	result.messgae = ""
	return &result
}
