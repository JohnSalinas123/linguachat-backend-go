package clerk

import (
	"fmt"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
)

/*
type clerkKeys struct {
	publicKey string
	privateKey string
}

var (
	keys *clerkKeys
)
*/

func InitialClerkSetup(secretKey string) error {
	if secretKey == "" {
		return fmt.Errorf("clerk secrey key is empty")
	}

	clerkSDK.SetKey(secretKey)
	return nil
}