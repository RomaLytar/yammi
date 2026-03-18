package integration

import "testing"

// --- Happy path (used in lifecycle chain) ---

func testRegister(t *testing.T) {
	resp, code, err := state.api.Register(state.email, state.password, state.name)
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	requireStatus(t, "Register", code, 201)
	requireNotEmpty(t, "user_id", resp.UserID)
	requireNotEmpty(t, "access_token", resp.AccessToken)
	requireNotEmpty(t, "refresh_token", resp.RefreshToken)

	state.userID = resp.UserID
	state.accessToken = resp.AccessToken
	state.refreshToken = resp.RefreshToken

	t.Logf("Registered user %s (%s)", state.userID, state.email)
}

// --- Negative cases ---

func testRegisterDuplicateEmail(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	_, code, _ := state.api.Register(state.email, state.password, "Duplicate User")
	requireStatus(t, "Register duplicate email", code, 409)

	t.Log("Duplicate email returns HTTP 409 Conflict")
}

func testRegisterEmptyEmail(t *testing.T) {
	_, code, _ := state.api.Register("", state.password, "No Email")
	requireStatus(t, "Register empty email", code, 400)

	t.Log("Empty email returns HTTP 400")
}

func testRegisterEmptyPassword(t *testing.T) {
	_, code, _ := state.api.Register("nopass@yammi.io", "", "No Pass")
	requireStatus(t, "Register empty password", code, 400)

	t.Log("Empty password returns HTTP 400")
}

func testRegisterWeakPassword(t *testing.T) {
	_, code, _ := state.api.Register("weak@yammi.io", "short", "Weak Pass")
	requireStatus(t, "Register weak password", code, 400)

	t.Log("Weak password (<8 chars) returns HTTP 400")
}

func testRegisterEmptyName(t *testing.T) {
	_, code, _ := state.api.Register("noname@yammi.io", "ValidPass123", "")
	requireStatus(t, "Register empty name", code, 400)

	t.Log("Empty name returns HTTP 400")
}
