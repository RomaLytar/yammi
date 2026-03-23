package board

import (
	"net/http"
	"time"
)

// HTTPMembershipChecker проверяет членство в доске через HTTP-запрос к API Gateway.
type HTTPMembershipChecker struct {
	baseURL string
	client  *http.Client
}

// NewHTTPMembershipChecker создаёт checker. baseURL — адрес API Gateway (например http://api-gateway:8080).
func NewHTTPMembershipChecker(baseURL string) *HTTPMembershipChecker {
	return &HTTPMembershipChecker{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// IsMember проверяет, является ли пользователь членом доски.
// Делает GET /api/v1/boards/{boardID} с Bearer-токеном пользователя.
// Если Board Service возвращает 200 — пользователь является участником.
func (c *HTTPMembershipChecker) IsMember(boardID, token string) bool {
	url := c.baseURL + "/api/v1/boards/" + boardID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
