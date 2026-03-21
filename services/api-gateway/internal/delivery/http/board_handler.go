package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	boardpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1/board"
)

type BoardHandler struct {
	client boardpb.BoardServiceClient
}

func NewBoardHandler(client boardpb.BoardServiceClient) *BoardHandler {
	return &BoardHandler{client: client}
}

// ============================================================================
// Board Handlers
// ============================================================================

// CreateBoard POST /api/v1/boards
func (h *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateBoard(r.Context(), &boardpb.CreateBoardRequest{
		Title:       req.Title,
		Description: req.Description,
		OwnerId:     userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"board": mapBoardFromProto(resp.Board),
	})
}

// GetBoard GET /api/v1/boards/{id}
func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	resp, err := h.client.GetBoard(r.Context(), &boardpb.GetBoardRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"board":   mapBoardFromProto(resp.Board),
		"columns": mapColumnsFromProto(resp.Columns),
		"members": mapMembersFromProto(resp.Members),
	})
}

// ListBoards GET /api/v1/boards
func (h *BoardHandler) ListBoards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQueryParam(r, "limit", 20)
	cursor := r.URL.Query().Get("cursor")
	ownerOnly := r.URL.Query().Get("owner_only") == "true"
	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sort_by")

	resp, err := h.client.ListBoards(r.Context(), &boardpb.ListBoardsRequest{
		UserId:    userID,
		Limit:     int32(limit),
		Cursor:    cursor,
		OwnerOnly: ownerOnly,
		Search:    search,
		SortBy:    sortBy,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"boards":      mapBoardsFromProto(resp.Boards),
		"next_cursor": resp.NextCursor,
	})
}

// UpdateBoard PUT /api/v1/boards/{id}
func (h *BoardHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateBoard(r.Context(), &boardpb.UpdateBoardRequest{
		BoardId:     boardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Version:     req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"board": mapBoardFromProto(resp.Board),
	})
}

// DeleteBoards POST /api/v1/boards/delete — удаление одной или нескольких досок
func (h *BoardHandler) DeleteBoards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		BoardIDs []string `json:"board_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.BoardIDs) == 0 {
		writeError(w, http.StatusBadRequest, "board_ids is required")
		return
	}

	_, err := h.client.DeleteBoard(r.Context(), &boardpb.DeleteBoardRequest{
		BoardIds: req.BoardIDs,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// ============================================================================
// Column Handlers
// ============================================================================

// AddColumn POST /api/v1/boards/{id}/columns
func (h *BoardHandler) AddColumn(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req struct {
		Title    string `json:"title"`
		Position int32  `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.AddColumn(r.Context(), &boardpb.AddColumnRequest{
		BoardId:  boardID,
		UserId:   userID,
		Title:    req.Title,
		Position: req.Position,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"column": mapColumnFromProto(resp.Column),
	})
}

// GetColumns GET /api/v1/boards/{id}/columns
func (h *BoardHandler) GetColumns(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	resp, err := h.client.GetColumns(r.Context(), &boardpb.GetColumnsRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"columns": mapColumnsFromProto(resp.Columns),
	})
}

// UpdateColumn PUT /api/v1/columns/{id}
func (h *BoardHandler) UpdateColumn(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	var req struct {
		BoardID string `json:"board_id"`
		Title   string `json:"title"`
		Version int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateColumn(r.Context(), &boardpb.UpdateColumnRequest{
		ColumnId: columnID,
		BoardId:  req.BoardID,
		UserId:   userID,
		Title:    req.Title,
		Version:  req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"column": mapColumnFromProto(resp.Column),
	})
}

// DeleteColumn DELETE /api/v1/columns/{id}
func (h *BoardHandler) DeleteColumn(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	var req struct {
		BoardID string `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.client.DeleteColumn(r.Context(), &boardpb.DeleteColumnRequest{
		ColumnId: columnID,
		BoardId:  req.BoardID,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// ReorderColumns PUT /api/v1/boards/{id}/columns/reorder
func (h *BoardHandler) ReorderColumns(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req struct {
		Positions []struct {
			ColumnID string `json:"column_id"`
			Position int32  `json:"position"`
		} `json:"positions"`
		Version int32 `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	positions := make([]*boardpb.ColumnPosition, len(req.Positions))
	for i, p := range req.Positions {
		positions[i] = &boardpb.ColumnPosition{
			ColumnId: p.ColumnID,
			Position: p.Position,
		}
	}

	resp, err := h.client.ReorderColumns(r.Context(), &boardpb.ReorderColumnsRequest{
		BoardId:   boardID,
		UserId:    userID,
		Positions: positions,
		Version:   req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"columns": mapColumnsFromProto(resp.Columns),
	})
}

// ============================================================================
// Card Handlers
// ============================================================================

// CreateCard POST /api/v1/columns/{id}/cards
func (h *BoardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	var req struct {
		BoardID     string `json:"board_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Position    string `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreateCard(r.Context(), &boardpb.CreateCardRequest{
		ColumnId:    columnID,
		BoardId:     req.BoardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Position:    req.Position,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// GetCard GET /api/v1/cards/{id}
func (h *BoardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.GetCard(r.Context(), &boardpb.GetCardRequest{
		CardId:  cardID,
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// GetCards GET /api/v1/columns/{id}/cards
func (h *BoardHandler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	columnID := r.PathValue("id")
	if columnID == "" {
		writeError(w, http.StatusBadRequest, "column id is required")
		return
	}

	boardID := r.URL.Query().Get("board_id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board_id query parameter is required")
		return
	}

	resp, err := h.client.GetCards(r.Context(), &boardpb.GetCardsRequest{
		ColumnId: columnID,
		BoardId:  boardID,
		UserId:   userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cards": mapCardsFromProto(resp.Cards),
	})
}

// UpdateCard PUT /api/v1/cards/{id}
func (h *BoardHandler) UpdateCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	var req struct {
		BoardID     string `json:"board_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		AssigneeID  string `json:"assignee_id"`
		Version     int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateCard(r.Context(), &boardpb.UpdateCardRequest{
		CardId:      cardID,
		BoardId:     req.BoardID,
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		AssigneeId:  req.AssigneeID,
		Version:     req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card": mapCardFromProto(resp.Card),
	})
}

// MoveCard PUT /api/v1/cards/{id}/move
func (h *BoardHandler) MoveCard(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		writeError(w, http.StatusBadRequest, "card id is required")
		return
	}

	var req struct {
		BoardID      string `json:"board_id"`
		FromColumnID string `json:"from_column_id"`
		ToColumnID   string `json:"to_column_id"`
		Position     string `json:"position"`
		Version      int32  `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.MoveCard(r.Context(), &boardpb.MoveCardRequest{
		CardId:       cardID,
		BoardId:      req.BoardID,
		FromColumnId: req.FromColumnID,
		ToColumnId:   req.ToColumnID,
		Position:     req.Position,
		UserId:       userID,
		Version:      req.Version,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"card":            mapCardFromProto(resp.Card),
		"cards_in_column": mapCardsFromProto(resp.CardsInColumn),
	})
}

// DeleteCards POST /api/v1/cards/delete — удаление одной или нескольких карточек
func (h *BoardHandler) DeleteCards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		CardIDs []string `json:"card_ids"`
		BoardID string   `json:"board_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.CardIDs) == 0 {
		writeError(w, http.StatusBadRequest, "card_ids is required")
		return
	}
	if req.BoardID == "" {
		writeError(w, http.StatusBadRequest, "board_id is required")
		return
	}

	_, err := h.client.DeleteCard(r.Context(), &boardpb.DeleteCardRequest{
		CardIds: req.CardIDs,
		BoardId: req.BoardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "deleted"})
}

// ============================================================================
// Member Handlers
// ============================================================================

// AddMember POST /api/v1/boards/{id}/members
func (h *BoardHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	var req struct {
		MemberUserID string `json:"user_id"`
		Role         string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.AddMember(r.Context(), &boardpb.AddMemberRequest{
		BoardId:      boardID,
		UserId:       userID,
		MemberUserId: req.MemberUserID,
		Role:         req.Role,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"member": mapMemberFromProto(resp.Member),
	})
}

// RemoveMember DELETE /api/v1/boards/{boardId}/members/{userId}
func (h *BoardHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("boardId")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	memberUserID := r.PathValue("userId")
	if memberUserID == "" {
		writeError(w, http.StatusBadRequest, "member user id is required")
		return
	}

	_, err := h.client.RemoveMember(r.Context(), &boardpb.RemoveMemberRequest{
		BoardId:      boardID,
		UserId:       userID,
		MemberUserId: memberUserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{Status: "removed"})
}

// ListMembers GET /api/v1/boards/{id}/members
func (h *BoardHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	boardID := r.PathValue("id")
	if boardID == "" {
		writeError(w, http.StatusBadRequest, "board id is required")
		return
	}

	resp, err := h.client.ListMembers(r.Context(), &boardpb.ListMembersRequest{
		BoardId: boardID,
		UserId:  userID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"members": mapMembersFromProto(resp.Members),
	})
}

// ============================================================================
// Helper Functions - Mappers (proto → JSON with snake_case)
// ============================================================================

func mapBoardFromProto(pb *boardpb.Board) boardResponse {
	return boardResponse{
		ID:          pb.Id,
		Title:       pb.Title,
		Description: pb.Description,
		OwnerID:     pb.OwnerId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapBoardsFromProto(pbs []*boardpb.Board) []boardResponse {
	boards := make([]boardResponse, len(pbs))
	for i, pb := range pbs {
		boards[i] = mapBoardFromProto(pb)
	}
	return boards
}

func mapColumnFromProto(pb *boardpb.Column) columnResponse {
	return columnResponse{
		ID:        pb.Id,
		BoardID:   pb.BoardId,
		Title:     pb.Title,
		Position:  pb.Position,
		Version:   pb.Version,
		CreatedAt: pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapColumnsFromProto(pbs []*boardpb.Column) []columnResponse {
	columns := make([]columnResponse, len(pbs))
	for i, pb := range pbs {
		columns[i] = mapColumnFromProto(pb)
	}
	return columns
}

func mapCardFromProto(pb *boardpb.Card) cardResponse {
	return cardResponse{
		ID:          pb.Id,
		ColumnID:    pb.ColumnId,
		BoardID:     pb.BoardId,
		Title:       pb.Title,
		Description: pb.Description,
		Position:    pb.Position,
		AssigneeID:  pb.AssigneeId,
		CreatorID:   pb.CreatorId,
		Version:     pb.Version,
		CreatedAt:   pb.CreatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   pb.UpdatedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapCardsFromProto(pbs []*boardpb.Card) []cardResponse {
	cards := make([]cardResponse, len(pbs))
	for i, pb := range pbs {
		cards[i] = mapCardFromProto(pb)
	}
	return cards
}

func mapMemberFromProto(pb *boardpb.BoardMember) memberResponse {
	return memberResponse{
		UserID:   pb.UserId,
		Role:     pb.Role,
		Version:  pb.Version,
		JoinedAt: pb.JoinedAt.AsTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapMembersFromProto(pbs []*boardpb.BoardMember) []memberResponse {
	members := make([]memberResponse, len(pbs))
	for i, pb := range pbs {
		members[i] = mapMemberFromProto(pb)
	}
	return members
}

func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return n
}
