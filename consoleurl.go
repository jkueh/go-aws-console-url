package awsconsoleurl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// A helper package to generate a console sign-in URL from a credentials Value object.

// Debug - Enables debug output
var Debug bool

// URLCredentials - The data structure that is embedded into the URL in the Session query parameter.
type URLCredentials struct {
	SessionID    string `json:"sessionId"`
	SessionKey   string `json:"sessionKey"`
	SessionToken string `json:"sessionToken"`
}

// SignInToken - The payload that's returned from the token request endpoint.
type SignInToken struct {
	Token string `json:"SigninToken"`
}

// GetSignInToken - Sends the initial request to retrieve the sign-in token.
func GetSignInToken(creds *credentials.Value) (*SignInToken, error) {
	urlCreds := URLCredentials{
		SessionID:    creds.AccessKeyID,
		SessionKey:   creds.SecretAccessKey,
		SessionToken: creds.SessionToken,
	}

	byteArr, err := json.Marshal(&urlCreds)
	if err != nil {
		log.Println("An error occurred during JSON marshaling of the existing credentials.")
		os.Exit(10)
	}

	tokenRequestEndpoint := fmt.Sprintf(
		"https://signin.aws.amazon.com/federation?Action=getSigninToken&Session=%s",
		url.QueryEscape(string(byteArr)),
	)

	tokenRequest, err := http.NewRequest(http.MethodGet, tokenRequestEndpoint, nil)
	if err != nil {
		if Debug {
			log.Println("Unable to build tokenRequest:", err)
		}
		return &SignInToken{}, err
	}

	tokenResponse, err := http.DefaultClient.Do(tokenRequest)
	if err != nil {
		if Debug {
			log.Println("An error occurred while making the request to the sign-in token endpoint:", err)
		}
		return &SignInToken{}, err
	}

	tokenResponseBody, err := ioutil.ReadAll(tokenResponse.Body)
	if err != nil {
		if Debug {
			log.Println("Unable to read token response body:", err)
		}
		return &SignInToken{}, err
	}

	// Unmarshal into the token struct
	var SIToken SignInToken
	err = json.Unmarshal(tokenResponseBody, &SIToken)
	if err != nil {
		if Debug {
			log.Println("Unable to unmarshal SignInToken:", err)
			log.Println("Response body:")
			fmt.Println(string(tokenResponseBody))
		}
		return &SignInToken{}, nil
	}

	return &SIToken, nil
}

// GetSignInURL - Returns the sign-in URL as a string.
func GetSignInURL(creds *credentials.Value) (string, error) {
	token, err := GetSignInToken(creds)
	return fmt.Sprintf(
		"https://signin.aws.amazon.com/federation?Action=login&Destination=%s&SigninToken=%s",
		url.QueryEscape("https://console.aws.amazon.com/"),
		token.Token,
	), err
}
