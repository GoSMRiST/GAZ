package grpc

import (
	"auth_users/internal/core"
	"auth_users/internal/core/auth"
	"context"
	"errors"
	token "github.com/GoSMRiST/protosGaz/gen/go/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JwtService interface {
	ParseJwtToken(tokenStr string) (*auth.Jwt, error)
}

type AuthGrpcServer struct {
	serv JwtService
	token.UnimplementedTokenServer
}

func RegisterAuthServer(gRPC *grpc.Server, serv JwtService) {
	token.RegisterTokenServer(gRPC, &AuthGrpcServer{
		serv: serv,
	})
}

func (as *AuthGrpcServer) ValidateToken(ctx context.Context,
	req *token.ValidateTokenRequest,
) (*token.ValidateTokenResponse, error) {
	jwtStruct, err := as.serv.ParseJwtToken(req.GetToken())
	if err != nil {
		if errors.Is(err, core.ErrTokenExpired) {
			return nil, status.Error(codes.Unauthenticated, "Token is expired")
		}

		if errors.Is(err, core.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "Token is invalid")
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &token.ValidateTokenResponse{
		UserId: (int64)(jwtStruct.UserID),
	}

	return resp, nil
}
