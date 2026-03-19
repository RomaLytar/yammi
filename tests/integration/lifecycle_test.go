package integration

import (
	"testing"
	"time"
)

// testDeleteUser deletes the user from Auth Service (cascading refresh_tokens).
// Auth publishes UserDeleted event to NATS for User Service to clean up.
func testDeleteUser(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	code, err := state.api.DeleteUser(state.userID)
	if err != nil {
		t.Fatalf("DeleteUser request failed: %v", err)
	}
	requireStatus(t, "DeleteUser", code, 200)

	t.Logf("User %s deleted from Auth Service", state.userID)
}

// testLoginAfterDelete proves the auth record is gone.
func testLoginAfterDelete(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	_, code, _ := state.api.Login(state.email, state.password)
	if code == 200 {
		t.Fatal("Login should fail after user deletion, but got 200")
	}

	t.Logf("Login after delete returns HTTP %d (user gone from Auth)", code)
}

// testProfileGoneAfterDelete polls User Service until the profile is deleted (async NATS delivery).
func testProfileGoneAfterDelete(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	state.api.WaitForProfileDeletion(t, state.userID, 5*time.Second)

	t.Logf("Profile removed from User Service via NATS UserDeleted event")
}

// testReRegisterSameEmail proves full cleanup: the email is available again
// and a new user gets a different ID.
func testReRegisterSameEmail(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	resp, code, err := state.api.Register(state.email, state.password, "Re-registered User")
	if err != nil {
		t.Fatalf("Re-register request failed: %v", err)
	}
	requireStatus(t, "Re-register", code, 201)
	requireNotEmpty(t, "new user_id", resp.UserID)
	requireNotEqual(t, "user_id differs from original", resp.UserID, state.userID)

	// Cleanup: delete the re-registered user (нужен его собственный токен)
	state.api.SetToken(resp.AccessToken)
	state.api.DeleteUser(resp.UserID)

	t.Logf("Same email re-registered (new user_id=%s) — full cleanup confirmed", resp.UserID)
}
