package grpc

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"auth_users/internal/core/auth/dto"
	"context"
	"errors"
	"github.com/GoSMRiST/protosGaz/gen/go/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService interface {
	GetUserByID(ctx context.Context) (*dto.GetByIdResponse, error)
	UpdateSubscription(ctx context.Context, sub *dto.UpdateSubscriptionRequest) error
}

type AuthGrpcServer struct {
	serv UserService
	user.UnimplementedUserServer
}

func RegisterUserServer(gRPC *grpc.Server, serv UserService) {
	user.RegisterUserServer(gRPC, &AuthGrpcServer{
		serv: serv,
	})
}

func (s *AuthGrpcServer) GetUser(ctx context.Context, userReq *user.GetUserRequest) (*user.GetUserResponse, error) {
	// Кладём user_id из запроса в контекст — именно его читает UserService.GetUserByID
	ctx = context.WithValue(ctx, auth.UserIDKey, int(userReq.GetUserId()))

	resp, err := s.serv.GetUserByID(ctx)
	if err != nil {
		if errors.Is(err, core.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "Invalid input")
		}

		if errors.Is(err, core.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &user.GetUserResponse{
		UserId:        int64(resp.ID),
		Nickname:      resp.Nickname,
		Email:         resp.Email,
		BirthDate:     timestamppb.New(resp.BirthDate),
		Gender:        resp.Gender,
		AvatarUrl:     resp.AvatarURL,
		MeetingsCount: int64(resp.MeetingsCount),
		Subscription:  resp.Subscription,
	}, nil
}

func (s *AuthGrpcServer) UpdateSubscription(ctx context.Context, userReq *user.UpdateSubscriptionRequest) (*user.UpdateSubscriptionResponse, error) {
	ctx = context.WithValue(ctx, auth.UserIDKey, int(userReq.GetUserId()))

	sub := &dto.UpdateSubscriptionRequest{
		SubscriptionStatus: userReq.GetSubscription(),
	}

	if err := s.serv.UpdateSubscription(ctx, sub); err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "Unauthenticated")
		}
		if errors.Is(err, core.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, "Invalid input")
		}
		if errors.Is(err, core.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}

		return nil, status.Error(codes.Internal, "Internal Error")
	}

	return &user.UpdateSubscriptionResponse{
		Subscription: userReq.GetSubscription(),
	}, nil
}
