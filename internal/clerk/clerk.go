package clerk

import (
	"fmt"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
	svix "github.com/svix/svix-webhooks/go"
)

var webhookVerifier *svix.Webhook

// intialClerkSetup returns error or nil
// intial setup of clerk, sets secret key for session validation
// and intializes svix secret to handle clerk webhook requests
func InitialClerkSetup(clerkSecret string, clerkWHSecret string) error {

	// set clerk secret
	if clerkSecret == "" {
		return fmt.Errorf("clerk secret secret is empty")
	}
	clerkSDK.SetKey(clerkSecret)


	// set svix secret
	if clerkWHSecret == "" {
		return fmt.Errorf(("clerk webhook secret is empty"))
	}
	wh, err := svix.NewWebhook(clerkWHSecret)
	if err != nil {
		return fmt.Errorf("failed to intialize webhook verifier: %w", err)
	}
	webhookVerifier = wh

	return nil
}