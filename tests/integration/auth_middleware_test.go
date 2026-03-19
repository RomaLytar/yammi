package integration

import (
	"fmt"
	"testing"
	"time"
)

// --- 401 Unauthorized: запросы без токена ---

func testGetProfileNoToken(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	_, code, _ := state.api.GetProfileNoAuth(state.userID)
	requireStatus(t, "GET profile no token", code, 401)

	t.Log("GET profile without token returns 401")
}

func testUpdateProfileNoToken(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	_, code, _ := state.api.UpdateProfileNoAuth(state.userID, "Hacker", "", "")
	requireStatus(t, "PUT profile no token", code, 401)

	t.Log("PUT profile without token returns 401")
}

func testDeleteUserNoToken(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	code, _ := state.api.DeleteUserNoAuth(state.userID)
	requireStatus(t, "DELETE user no token", code, 401)

	t.Log("DELETE user without token returns 401")
}

func testRefreshNoToken(t *testing.T) {
	if state.refreshToken == "" {
		t.Skip("skipped: no refresh_token")
	}

	_, code, _ := state.api.RefreshTokenNoAuth(state.refreshToken)
	requireStatus(t, "POST refresh no token", code, 401)

	t.Log("POST refresh without access token returns 401")
}

func testRevokeNoToken(t *testing.T) {
	if state.refreshToken == "" {
		t.Skip("skipped: no refresh_token")
	}

	code, _ := state.api.RevokeTokenNoAuth(state.refreshToken)
	requireStatus(t, "POST revoke no token", code, 401)

	t.Log("POST revoke without access token returns 401")
}

// --- 403 Forbidden: чужой аккаунт ---

func testUpdateProfileForbidden(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	// Регистрируем второго юзера
	email2 := fmt.Sprintf("forbidden-test-%d@yammi.io", time.Now().UnixNano())
	resp2, code, err := state.api.Register(email2, "TestPassword123", "Victim User")
	if err != nil {
		t.Fatalf("Register second user failed: %v", err)
	}
	requireStatus(t, "Register victim", code, 201)

	// User1 пытается обновить профиль User2
	_, code, _ = state.api.UpdateProfileAs(resp2.UserID, state.accessToken, "Hacked Name", "", "pwned")
	requireStatus(t, "PUT other user's profile", code, 403)

	// Cleanup: удаляем второго юзера его собственным токеном
	state.api.DeleteUserAs(resp2.UserID, resp2.AccessToken)

	t.Log("Updating another user's profile returns 403")
}

func testDeleteUserForbidden(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	// Регистрируем второго юзера
	email2 := fmt.Sprintf("forbidden-del-%d@yammi.io", time.Now().UnixNano())
	resp2, code, err := state.api.Register(email2, "TestPassword123", "Victim User 2")
	if err != nil {
		t.Fatalf("Register second user failed: %v", err)
	}
	requireStatus(t, "Register victim", code, 201)

	// User1 пытается удалить User2
	code, _ = state.api.DeleteUserAs(resp2.UserID, state.accessToken)
	requireStatus(t, "DELETE other user", code, 403)

	// Проверяем что User2 всё ещё жив
	_, code, _ = state.api.GetProfileAs(resp2.UserID, resp2.AccessToken)
	requireStatus(t, "Victim still alive", code, 200)

	// Cleanup
	state.api.DeleteUserAs(resp2.UserID, resp2.AccessToken)

	t.Log("Deleting another user returns 403, victim survives")
}

// --- 200 OK: авторизованный юзер видит чужой профиль ---

func testViewOtherProfileAllowed(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	// Регистрируем второго юзера
	email2 := fmt.Sprintf("viewtest-%d@yammi.io", time.Now().UnixNano())
	resp2, code, err := state.api.Register(email2, "TestPassword123", "Viewable User")
	if err != nil {
		t.Fatalf("Register second user failed: %v", err)
	}
	requireStatus(t, "Register viewable", code, 201)

	// Ждём создания профиля через NATS
	time.Sleep(2 * time.Second)

	// User1 смотрит профиль User2 — должно работать
	profile, code, _ := state.api.GetProfileAs(resp2.UserID, state.accessToken)
	requireStatus(t, "View other user profile", code, 200)
	requireEqual(t, "other user email", profile.Email, email2)

	// Cleanup
	state.api.DeleteUserAs(resp2.UserID, resp2.AccessToken)

	t.Log("Authenticated user can view another user's profile")
}

// --- Invalid token ---

func testGetProfileInvalidToken(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	_, code, _ := state.api.GetProfileAs(state.userID, "totally.invalid.token")
	requireStatus(t, "GET profile invalid token", code, 401)

	t.Log("Invalid token returns 401")
}
