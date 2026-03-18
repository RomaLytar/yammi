package integration

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// testState carries data between sequential subtests across all test files.
type testState struct {
	api          *APIClient
	email        string
	password     string
	name         string
	userID       string
	accessToken  string
	refreshToken string
}

var state testState

func TestMain(m *testing.M) {
	baseURL := os.Getenv("API_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	state.api = NewAPIClient(baseURL)

	fmt.Printf("Waiting for API Gateway at %s...\n", baseURL)
	if err := state.api.WaitForHealthy(30 * time.Second); err != nil {
		fmt.Printf("FAIL: %v\nMake sure services are running: docker compose up -d\n", err)
		os.Exit(1)
	}
	fmt.Println("API Gateway is healthy")

	os.Exit(m.Run())
}

// TestUserLifecycle runs the full user lifecycle as sequential subtests.
// The chain: register → verify → use → revoke → recover → rotate → delete → verify cleanup.
func TestUserLifecycle(t *testing.T) {
	state.email = fmt.Sprintf("inttest-%d@yammi.io", time.Now().UnixNano())
	state.password = "TestPassword123"
	state.name = "Integration Test User"

	// --- Registration ---
	t.Run("01_Register", testRegister)
	t.Run("02_Register_DuplicateEmail", testRegisterDuplicateEmail)
	t.Run("03_Register_EmptyEmail", testRegisterEmptyEmail)
	t.Run("04_Register_EmptyPassword", testRegisterEmptyPassword)
	t.Run("05_Register_WeakPassword", testRegisterWeakPassword)
	t.Run("06_Register_EmptyName", testRegisterEmptyName)

	// --- Login ---
	t.Run("07_Login", testLogin)
	t.Run("08_Login_WrongPassword", testLoginWrongPassword)
	t.Run("09_Login_NonExistentUser", testLoginNonExistentUser)
	t.Run("10_Login_EmptyFields", testLoginEmptyFields)

	// --- Profile (async via NATS) ---
	t.Run("11_Profile_CreatedViaNATS", testProfileCreatedViaNATS)
	t.Run("12_Profile_Update", testUpdateProfile)
	t.Run("13_Profile_GetNonExistent", testGetProfileNonExistent)
	t.Run("14_Profile_UpdateEmptyName", testUpdateProfileEmptyName)
	t.Run("15_Profile_UpdateNonExistent", testUpdateProfileNonExistent)
	t.Run("16_Profile_DeleteNonExistent", testDeleteNonExistentUser)

	// --- Tokens ---
	t.Run("17_Token_Revoke_RefreshFails", testRevokeToken)
	t.Run("18_Token_LoginAfterRevoke", testLoginAfterRevoke)
	t.Run("19_Token_Rotation", testRefreshTokenRotation)
	t.Run("20_Token_RefreshInvalid", testRefreshWithInvalidToken)
	t.Run("21_Token_RefreshEmpty", testRefreshWithEmptyToken)
	t.Run("22_Token_RevokeAlreadyRevoked", testRevokeAlreadyRevoked)
	t.Run("23_Token_DoubleRefreshReplay", testDoubleRefreshReplay)

	// --- Deletion & cleanup ---
	t.Run("24_Delete_User", testDeleteUser)
	t.Run("25_Delete_LoginFails", testLoginAfterDelete)
	t.Run("26_Delete_ProfileGoneViaNATS", testProfileGoneAfterDelete)
	t.Run("27_Delete_ReRegisterSameEmail", testReRegisterSameEmail)
}
