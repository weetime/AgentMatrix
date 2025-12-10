package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	// AuthorizationHeader Authorization请求头
	AuthorizationHeader = "Authorization"
)

// contextKey 用于Context的key类型，避免使用string作为key
type contextKey string

const (
	// UserIDKey Context中存储用户ID的key
	UserIDKey contextKey = "user_id"
	// UserDetailKey Context中存储用户详情的key
	UserDetailKey contextKey = "user_detail"
)

// UserDetail 用户详情（对应Java的UserDetail）
type UserDetail struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	SuperAdmin int32  `json:"super_admin"`
	Status     int32  `json:"status"`
	Token      string `json:"token"`
}

// TokenService Token服务接口
type TokenService interface {
	GetUserByToken(ctx context.Context, token string) (*UserDetail, error)
}

// isPathAllowed 检查路径是否在白名单中（不需要认证）
// 参考Java的ShiroConfig中的anon路径配置
func isPathAllowed(path string) bool {
	// 定义不需要认证的路径白名单（参考Java的ShiroConfig）
	allowedPaths := []string{
		"/q/",                           // Swagger UI
		"/v3/api-docs/",                 // OpenAPI文档
		"/doc.html",                     // Swagger UI页面
		"/favicon.ico",                  // 图标
		"/user/login",                   // 登录
		"/user/register",                // 注册
		"/user/pub-config",              // 公共配置
		"/user/captcha",                 // 验证码
		"/user/smsVerification",         // 短信验证
		"/user/retrieve-password",       // 找回密码
		"/agent/chat-history/download/", // 聊天记录下载
		"/agent/play/",                  // 智能体播放
		"/voiceClone/play/",             // 声音克隆播放
		"/ota/",                         // OTA相关
		"/otaMag/download/",             // OTA下载
		"/webjars/",                     // WebJars资源
		"/druid/",                       // Druid监控
		"/ws",                           // WebSocket
	}

	// 检查路径是否匹配白名单
	for _, allowedPath := range allowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			return true
		}
	}

	return false
}

// AuthMiddleware 认证中间件（对应Java的Oauth2Filter）
// 参考Java实现：如果没有token返回401，如果有token则验证并设置用户信息到context
func AuthMiddleware(tokenService TokenService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 从HTTP请求中获取请求对象
			httpReq, ok := kratoshttp.RequestFromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			// 对OPTIONS请求放行（CORS预检请求）
			if httpReq.Method == http.MethodOptions {
				return handler(ctx, req)
			}

			// 检查路径是否在白名单中（不需要认证）
			path := httpReq.URL.Path
			if isPathAllowed(path) {
				return handler(ctx, req)
			}

			// 提取Token
			token := extractTokenFromRequest(httpReq)

			// 如果没有token，返回401（参考Java的onAccessDenied逻辑）
			if token == "" {
				return nil, sendUnauthorizedResponse(ctx, 401, "未授权")
			}

			// 验证Token并获取用户信息
			user, err := tokenService.GetUserByToken(ctx, token)
			if err != nil || user == nil {
				// Token无效或过期，返回401
				return nil, sendUnauthorizedResponse(ctx, 401, "未授权")
			}

			// 检查账号状态（参考Java的Oauth2Realm逻辑）
			// status为0表示账号被锁定
			if user.Status == 0 {
				return nil, sendUnauthorizedResponse(ctx, 401, "账号已被锁定")
			}

			// 将用户信息存储到Context中
			ctx = context.WithValue(ctx, UserIDKey, user.ID)
			ctx = context.WithValue(ctx, UserDetailKey, user)

			return handler(ctx, req)
		}
	}
}

// extractTokenFromRequest 从HTTP请求中提取Token
// 参考Java实现：从Authorization header中提取Bearer Token
func extractTokenFromRequest(req *kratoshttp.Request) string {
	auth := req.Header.Get(AuthorizationHeader)
	if auth == "" {
		return ""
	}

	// 检查Bearer格式：Authorization: Bearer {token}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// sendUnauthorizedResponse 发送未授权响应（参考Java的Oauth2Filter实现）
// 返回错误，kratos会自动处理为HTTP响应
func sendUnauthorizedResponse(ctx context.Context, code int, msg string) error {
	// 创建未授权错误，kratos会自动转换为HTTP 401响应
	// 注意：kratos的错误处理会自动设置响应格式，但我们需要确保CORS头被设置
	return errors.Unauthorized("UNAUTHORIZED", msg)
}

// GetUserFromContext 从Context获取用户信息（对应Java的SecurityUser.getUser()）
func GetUserFromContext(ctx context.Context) (*UserDetail, error) {
	user, ok := ctx.Value(UserDetailKey).(*UserDetail)
	if !ok || user == nil {
		return nil, errors.Unauthorized("UNAUTHORIZED", "user not authenticated")
	}
	return user, nil
}

// GetUserIdFromContext 从Context获取用户ID（对应Java的SecurityUser.getUserId()）
func GetUserIdFromContext(ctx context.Context) (int64, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GetTokenFromContext 从Context获取Token（对应Java的SecurityUser.getToken()）
func GetTokenFromContext(ctx context.Context) (string, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return "", err
	}
	return user.Token, nil
}

func IsSuperAdmin(ctx context.Context) bool {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return false
	}
	return user.SuperAdmin == 1
}
