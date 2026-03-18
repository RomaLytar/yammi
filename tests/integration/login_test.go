package integration

import "testing"

// --- Happy path (used in lifecycle chain) ---

func testLogin(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	resp, code, err := state.api.Login(state.email, state.password)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	requireStatus(t, "Login", code, 200)
	requireEqual(t, "user_id", resp.UserID, state.userID)
	requireNotEmpty(t, "access_token", resp.AccessToken)
	requireNotEmpty(t, "refresh_token", resp.RefreshToken)

	state.accessToken = resp.AccessToken
	state.refreshToken = resp.RefreshToken

	t.Log("Login successful, new tokens issued")
}

func testLoginAfterRevoke(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	resp, code, err := state.api.Login(state.email, state.password)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	requireStatus(t, "Login after revoke", code, 200)
	requireEqual(t, "user_id", resp.UserID, state.userID)

	state.accessToken = resp.AccessToken
	state.refreshToken = resp.RefreshToken

	t.Log("Login after revoke successful — account intact")
}

// --- Negative cases ---

func testLoginWrongPassword(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	_, code, _ := state.api.Login(state.email, "TotallyWrongPassword")
	requireStatus(t, "Login wrong password", code, 401)

	t.Log("Wrong password returns HTTP 401")
}

func testLoginNonExistentUser(t *testing.T) {
	_, code, _ := state.api.Login("nobody-exists-here@yammi.io", "SomePassword1")
	requireStatus(t, "Login non-existent user", code, 404)

	t.Log("Non-existent user returns HTTP 404")
}

func testLoginEmptyFields(t *testing.T) {
	_, code, _ := state.api.Login("", "")
	requireStatus(t, "Login empty fields", code, 400)

	t.Log("Empty login fields returns HTTP 400")
}
