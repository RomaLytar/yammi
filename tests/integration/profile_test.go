package integration

import (
	"testing"
	"time"
)

// --- Happy path (used in lifecycle chain) ---

func testProfileCreatedViaNATS(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id from Register")
	}

	profile := state.api.WaitForProfile(t, state.userID, 5*time.Second)

	requireEqual(t, "profile.email", profile.Email, state.email)
	requireEqual(t, "profile.name", profile.Name, state.name)
	requireNotEmpty(t, "profile.created_at", profile.CreatedAt)

	t.Logf("Profile created in User Service via NATS (email=%s)", profile.Email)
}

func testUpdateProfile(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	newName := "Updated Integration User"
	newBio := "Testing Yammi!"
	newAvatar := "https://example.com/test-avatar.png"

	resp, code, err := state.api.UpdateProfile(state.userID, newName, newAvatar, newBio)
	if err != nil {
		t.Fatalf("UpdateProfile request failed: %v", err)
	}
	requireStatus(t, "UpdateProfile", code, 200)
	requireEqual(t, "name", resp.Name, newName)
	requireEqual(t, "bio", resp.Bio, newBio)
	requireEqual(t, "avatar_url", resp.AvatarURL, newAvatar)

	// Verify persistence with a separate GET
	profile, code, _ := state.api.GetProfile(state.userID)
	requireStatus(t, "GetProfile after update", code, 200)
	requireEqual(t, "persisted name", profile.Name, newName)
	requireEqual(t, "persisted bio", profile.Bio, newBio)
	requireEqual(t, "persisted avatar_url", profile.AvatarURL, newAvatar)

	t.Logf("Profile updated and persisted (name=%s, bio=%s)", newName, newBio)
}

// --- Negative cases ---

func testGetProfileNonExistent(t *testing.T) {
	_, code, _ := state.api.GetProfile("00000000-0000-0000-0000-000000000000")
	requireStatus(t, "GetProfile non-existent", code, 404)

	t.Log("Non-existent profile returns HTTP 404")
}

func testUpdateProfileEmptyName(t *testing.T) {
	if state.userID == "" {
		t.Skip("skipped: no user_id")
	}

	_, code, _ := state.api.UpdateProfile(state.userID, "", "", "")
	requireStatus(t, "UpdateProfile empty name", code, 400)

	t.Log("Update with empty name returns HTTP 400")
}

func testUpdateProfileNonExistent(t *testing.T) {
	// OwnerOnly middleware отклонит — чужой ID ≠ token.user_id → 403
	_, code, _ := state.api.UpdateProfile("00000000-0000-0000-0000-000000000000", "Ghost", "", "")
	requireStatus(t, "UpdateProfile non-existent (forbidden)", code, 403)

	t.Log("Update non-existent profile returns HTTP 403 (owner check first)")
}

func testDeleteNonExistentUser(t *testing.T) {
	// OwnerOnly middleware отклонит — чужой ID ≠ token.user_id → 403
	code, _ := state.api.DeleteUser("00000000-0000-0000-0000-000000000000")
	requireStatus(t, "Delete non-existent user (forbidden)", code, 403)

	t.Log("Delete non-existent user returns HTTP 403 (owner check first)")
}
