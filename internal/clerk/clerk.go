package clerk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
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


// updateUserPublicData updates a user's public data
func UpdateUserPublicData(fieldName string, fieldValue string, userID string) error {

	UserLangCode := map[string]interface{}{
		fieldName : fieldValue,
	}

	publicMetadataJSON, err := json.Marshal(UserLangCode)
	if err != nil {
		return fmt.Errorf("error marshaling metadata")
	}

	rawMetadata := json.RawMessage(publicMetadataJSON)

	user, err := user.UpdateMetadata(context.Background(), userID, &user.UpdateMetadataParams{
		PublicMetadata: &rawMetadata,
	})
	if err != nil {
		return fmt.Errorf("error updating metadata")
	}

	log.Printf("Successfully updated user metadata: %+v", user.PublicMetadata)

	return nil

}