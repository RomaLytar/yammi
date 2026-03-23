package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	boardpb "github.com/RomaLytar/yammi/services/board/api/proto/v1"
	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// CreateAutomationRule создает новое правило автоматизации (только owner)
func (s *BoardServiceServer) CreateAutomationRule(ctx context.Context, req *boardpb.CreateAutomationRuleRequest) (*boardpb.CreateAutomationRuleResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetTriggerType() == "" {
		return nil, status.Error(codes.InvalidArgument, "trigger_type is required")
	}
	if req.GetActionType() == "" {
		return nil, status.Error(codes.InvalidArgument, "action_type is required")
	}

	rule, err := s.createAutomationRule.Execute(ctx,
		req.GetBoardId(), req.GetUserId(), req.GetName(),
		domain.TriggerType(req.GetTriggerType()), req.GetTriggerConfig(),
		domain.ActionType(req.GetActionType()), req.GetActionConfig(),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.CreateAutomationRuleResponse{
		Rule: mapAutomationRuleToProto(rule),
	}, nil
}

// ListAutomationRules возвращает все правила автоматизации доски
func (s *BoardServiceServer) ListAutomationRules(ctx context.Context, req *boardpb.ListAutomationRulesRequest) (*boardpb.ListAutomationRulesResponse, error) {
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	rules, err := s.listAutomationRules.Execute(ctx, req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.ListAutomationRulesResponse{
		Rules: mapAutomationRulesToProto(rules),
	}, nil
}

// UpdateAutomationRule обновляет правило автоматизации (только owner)
func (s *BoardServiceServer) UpdateAutomationRule(ctx context.Context, req *boardpb.UpdateAutomationRuleRequest) (*boardpb.UpdateAutomationRuleResponse, error) {
	if req.GetRuleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rule_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	rule, err := s.updateAutomationRule.Execute(ctx,
		req.GetRuleId(), req.GetBoardId(), req.GetUserId(),
		req.GetName(), req.GetEnabled(),
		req.GetTriggerConfig(), req.GetActionConfig(),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.UpdateAutomationRuleResponse{
		Rule: mapAutomationRuleToProto(rule),
	}, nil
}

// DeleteAutomationRule удаляет правило автоматизации (только owner)
func (s *BoardServiceServer) DeleteAutomationRule(ctx context.Context, req *boardpb.DeleteAutomationRuleRequest) (*emptypb.Empty, error) {
	if req.GetRuleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rule_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.deleteAutomationRule.Execute(ctx, req.GetRuleId(), req.GetBoardId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &emptypb.Empty{}, nil
}

// GetAutomationHistory возвращает историю выполнений правила
func (s *BoardServiceServer) GetAutomationHistory(ctx context.Context, req *boardpb.GetAutomationHistoryRequest) (*boardpb.GetAutomationHistoryResponse, error) {
	if req.GetRuleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "rule_id is required")
	}
	if req.GetBoardId() == "" {
		return nil, status.Error(codes.InvalidArgument, "board_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}

	executions, err := s.getAutomationHistory.Execute(ctx, req.GetRuleId(), req.GetBoardId(), req.GetUserId(), limit)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &boardpb.GetAutomationHistoryResponse{
		Executions: mapAutomationExecutionsToProto(executions),
	}, nil
}
