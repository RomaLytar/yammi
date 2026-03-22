package grpc

import (
	commentpb "github.com/RomaLytar/yammi/services/comment/api/proto/v1"
	"github.com/RomaLytar/yammi/services/comment/internal/usecase"
)

// CommentServiceServer реализует gRPC сервер Comment Service
type CommentServiceServer struct {
	commentpb.UnimplementedCommentServiceServer

	createComment   *usecase.CreateCommentUseCase
	listComments    *usecase.ListCommentsUseCase
	updateComment   *usecase.UpdateCommentUseCase
	deleteComment   *usecase.DeleteCommentUseCase
	getCommentCount *usecase.GetCommentCountUseCase
}

// NewCommentServiceServer создает новый gRPC сервер с внедренными use cases
func NewCommentServiceServer(
	createComment *usecase.CreateCommentUseCase,
	listComments *usecase.ListCommentsUseCase,
	updateComment *usecase.UpdateCommentUseCase,
	deleteComment *usecase.DeleteCommentUseCase,
	getCommentCount *usecase.GetCommentCountUseCase,
) *CommentServiceServer {
	return &CommentServiceServer{
		createComment:   createComment,
		listComments:    listComments,
		updateComment:   updateComment,
		deleteComment:   deleteComment,
		getCommentCount: getCommentCount,
	}
}
