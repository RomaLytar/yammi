package integration

import "testing"

// --- Happy path (used in lifecycle chain) ---

func testRevokeToken(t *testing.T) {
	if state.refreshToken == "" {
		t.Skip("skipped: no refresh_token")
	}

	code, err := state.api.RevokeToken(state.refreshToken)
	if err != nil {
		t.Fatalf("RevokeToken request failed: %v", err)
	}
	requireStatus(t, "RevokeToken", code, 200)

	// Refresh with revoked token must fail
	_, code, _ = state.api.RefreshToken(state.refreshToken)
	if code == 200 {
		t.Fatal("RefreshToken should fail after revoke, but got 200")
	}

	t.Logf("Token revoked, refresh returns HTTP %d", code)
}

func testRefreshTokenRotation(t *testing.T) {
	if state.refreshToken == "" {
		t.Skip("skipped: no refresh_token")
	}

	oldRefresh := state.refreshToken
	oldAccess := state.accessToken

	resp, code, err := state.api.RefreshToken(oldRefresh)
	if err != nil {
		t.Fatalf("RefreshToken request failed: %v", err)
	}
	requireStatus(t, "RefreshToken", code, 200)
	requireNotEmpty(t, "new access_token", resp.AccessToken)
	requireNotEmpty(t, "new refresh_token", resp.RefreshToken)

	requireNotEqual(t, "access_token rotated", resp.AccessToken, oldAccess)
	requireNotEqual(t, "refresh_token rotated", resp.RefreshToken, oldRefresh)

	state.accessToken = resp.AccessToken
	state.refreshToken = resp.RefreshToken

	// Old refresh token must be invalid
	_, code, _ = state.api.RefreshToken(oldRefresh)
	if code == 200 {
		t.Fatal("Old refresh token should be invalid after rotation, but got 200")
	}

	t.Logf("Token rotation works: old refresh rejected (HTTP %d)", code)
}

// --- Negative cases ---

func testRefreshWithInvalidToken(t *testing.T) {
	_, code, _ := state.api.RefreshToken("00000000-0000-0000-0000-000000000000")
	if code == 200 {
		t.Fatal("Refresh with fake token should fail, but got 200")
	}

	t.Logf("Refresh with invalid token returns HTTP %d", code)
}

func testRefreshWithEmptyToken(t *testing.T) {
	_, code, _ := state.api.RefreshToken("")
	requireStatus(t, "Refresh empty token", code, 400)

	t.Log("Refresh with empty token returns HTTP 400")
}

func testRevokeAlreadyRevoked(t *testing.T) {
	if state.refreshToken == "" {
		t.Skip("skipped: no refresh_token")
	}

	// Revoke current token
	state.api.RevokeToken(state.refreshToken)

	// Try revoking again — should not crash (idempotent or error)
	code, _ := state.api.RevokeToken(state.refreshToken)
	if code == 500 {
		t.Fatalf("Revoking already revoked token should not return 500, got %d", code)
	}

	t.Logf("Revoking already revoked token returns HTTP %d (not 500)", code)
}

func testDoubleRefreshReplay(t *testing.T) {
	if state.email == "" {
		t.Skip("skipped: no email")
	}

	// Fresh login to get a clean token
	resp, code, _ := state.api.Login(state.email, state.password)
	requireStatus(t, "Login for double refresh", code, 200)

	// First refresh — should succeed
	newResp, code, _ := state.api.RefreshToken(resp.RefreshToken)
	requireStatus(t, "First refresh", code, 200)

	// Second refresh with SAME old token — must fail (replay attack protection)
	_, code, _ = state.api.RefreshToken(resp.RefreshToken)
	if code == 200 {
		t.Fatal("Replay of old refresh token should fail, but got 200")
	}

	state.accessToken = newResp.AccessToken
	state.refreshToken = newResp.RefreshToken

	t.Logf("Double refresh replay rejected (HTTP %d)", code)
}
