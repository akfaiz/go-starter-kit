package routes

import (
	"fmt"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler/dto"
	"github.com/labstack/echo/v5"
	"github.com/oaswrap/spec/adapter/echov5openapi"
	"github.com/oaswrap/spec/option"
	"go.uber.org/fx"
)

type RouteConfig struct {
	fx.In

	Echo   *echo.Echo
	Config config.Config

	AuthMiddleware echo.MiddlewareFunc `name:"auth"`

	AuthHandler        *handler.AuthHandler
	ProfileHandler     *handler.ProfileHandler
	UserHandler        *handler.UserHandler
	HealthCheckHandler *handler.HealthCheckHandler
}

func Register(rc RouteConfig) echov5openapi.Generator {
	rc.Echo.GET("/health", rc.HealthCheckHandler.HealthCheck)
	r := echov5openapi.NewRouter(
		rc.Echo,
		option.WithTitle("Go Starter Kit API"),
		option.WithVersion("1.0.0"),
		option.WithDescription("API starterkit with Echo v5, Gorm, Migris, and OTP auth reset flow"),
		option.WithReflectorConfig(option.StripDefNamePrefix("Dto")),
		option.WithSecurity("bearerAuth", option.SecurityHTTPBearer("Bearer")),
		option.WithServer(
			fmt.Sprintf("http://localhost:%d", rc.Config.Server.Port),
			option.ServerDescription("Local server"),
		),
		option.WithOpenAPIVersion("3.2.0"),
		option.WithScalar(),
	)

	v1 := r.Group("/api/v1")
	registerAuthRoutes(v1, rc)
	registerProfileRoutes(v1, rc)
	registerUserRoutes(v1, rc)

	return r
}

func registerAuthRoutes(v1 echov5openapi.Router, rc RouteConfig) {
	auth := v1.Group("/auth").With(option.GroupTags("Authentication"))
	auth.POST("/register", rc.AuthHandler.Register).With(
		option.Summary("User Registration"),
		option.Request(new(dto.RegisterRequest)),
		option.Response(201, new(dto.Response[dto.TokenResponse])),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	auth.POST("/login", rc.AuthHandler.Login).With(
		option.Summary("User Login"),
		option.Request(new(dto.LoginRequest)),
		option.Response(200, new(dto.Response[dto.TokenResponse])),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	auth.POST("/refresh-token", rc.AuthHandler.RefreshToken).With(
		option.Summary("Refresh Token"),
		option.Request(new(dto.RefreshTokenRequest)),
		option.Response(200, new(dto.Response[dto.TokenResponse])),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	auth.POST("/forgot-password/send-otp", rc.AuthHandler.SendForgotPasswordOTP).With(
		option.Summary("Send forgot password OTP"),
		option.Request(new(dto.SendForgotPasswordOTPRequest)),
		option.Response(200, new(dto.MessageResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	auth.POST("/forgot-password/verify-otp", rc.AuthHandler.VerifyForgotPasswordOTP).With(
		option.Summary("Verify forgot password OTP"),
		option.Request(new(dto.VerifyForgotPasswordOTPRequest)),
		option.Response(200, new(dto.MessageResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	auth.POST("/forgot-password/reset-password", rc.AuthHandler.ResetPasswordWithOTP).With(
		option.Summary("Reset password with OTP"),
		option.Request(new(dto.ResetPasswordWithOTPRequest)),
		option.Response(200, new(dto.MessageResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
}

func registerProfileRoutes(v1 echov5openapi.Router, rc RouteConfig) {
	profile := v1.Group("/profile", rc.AuthMiddleware).With(
		option.GroupTags("Profile"),
		option.GroupSecurity("bearerAuth"),
	)
	profile.GET("", rc.ProfileHandler.GetProfile).With(
		option.Summary("Get profile"),
		option.Response(200, new(dto.Response[dto.ProfileResponse])),
	)
	profile.PUT("", rc.ProfileHandler.UpdateProfile).With(
		option.Summary("Update profile"),
		option.Request(new(dto.UpdateProfileRequest)),
		option.Response(200, new(dto.Response[dto.ProfileResponse])),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	profile.PUT("/password", rc.ProfileHandler.ChangePassword).With(
		option.Summary("Update password"),
		option.Request(new(dto.ChangePasswordRequest)),
		option.Response(200, new(dto.MessageResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
	profile.DELETE("", rc.ProfileHandler.DeleteProfile).With(
		option.Summary("Delete profile"),
		option.Request(new(dto.DeleteProfileRequest)),
		option.Response(200, new(dto.MessageResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(422, new(dto.ValidationErrorResponse)),
	)
}

func registerUserRoutes(v1 echov5openapi.Router, rc RouteConfig) {
	users := v1.Group("/users", rc.AuthMiddleware).With(
		option.GroupTags("Users"),
		option.GroupSecurity("bearerAuth"),
	)
	users.GET("", rc.UserHandler.ListUsers).With(
		option.Summary("List users"),
		option.Request(new(dto.UserListRequest)),
		option.Response(200, new(dto.PaginatedResponse[*dto.UserResponse])),
		option.Response(401, new(dto.ErrorResponse)),
	)
	users.GET("/:id", rc.UserHandler.GetUser).With(
		option.Summary("Get user by ID"),
		option.Request(new(dto.UserGetRequest)),
		option.Response(200, new(dto.Response[dto.UserResponse])),
		option.Response(400, new(dto.ErrorResponse)),
		option.Response(401, new(dto.ErrorResponse)),
		option.Response(404, new(dto.ErrorResponse)),
	)
}
