package middleware

import (
	"context"
	"fmt"
	"g_dev/internal/auth"
	"net/http"
	"strings"
)

// JWT 토큰 기반 인증 처리
// 요청에서 JWT토큰을 추출하고 검증하여 사용자 정보를 컨텍스트에 추가합니다.
type JWTMiddleware struct {
	jwtAuth *auth.JWTAuth
}

// 컨텍스트에서 데이터를 저장하는 키
type ContextKey string

const (
	// 사용자 정보를 저장하는 키
	UserContextKey ContextKey = "user"
	// 토큰을 저장하는 키
	TokenContextKey ContextKey = "token"
)

// 사용자 정보를 담는 구조체
type UserInfo struct {
	UserID   uint
	Username string
	Role     string
	Token    string
}

// 새로운 JWT 미들웨어 인스턴스를 생성
func NewJWTMiddleware(jwtAuth *auth.JWTAuth) *JWTMiddleware {
	return &JWTMiddleware{
		jwtAuth: jwtAuth,
	}
}

// 토큰 검증, 사용자 정보를 컨텍스트에 추가
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization 헤더에서 토큰 추출
		token, err := m.extractToken(r)
		if err != nil {
			m.writeUnauthorizedResponse(w, "토큰이 필요합니다.")
			return
		}

		// JWT 토큰 검증
		claims, err := m.jwtAuth.ValidateAccessToken(token)
		if err != nil {
			m.writeUnauthorizedResponse(w, "유효하지 않은 토큰입니다.")
			return
		}

		// 사용자 정보를 컨텍스트에 추가
		userInfo := &UserInfo{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
			Token:    token,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		ctx = context.WithValue(ctx, TokenContextKey, token)

		// 다음 핸들러 호출
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 특정 역할을 가진 사용자만 접근을 허용
// 역할이 일치하지 않는 경우 403 Forbidden 반환
func (m *JWTMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 사용자 정보 가져옴
			userInfo, ok := r.Context().Value(UserContextKey).(*UserInfo)
			if !ok {
				m.writeForbiddenResponse(w, "사용자 정보를 찾을 수 없습니다.")
				return
			}

			// 역할 확인
			if userInfo.Role != requiredRole {
				m.writeForbiddenResponse(w, fmt.Sprintf("'%s' 역할이 필요합니다.", requiredRole))
				return
			}

			// 다음 핸들러 호출
			next.ServeHTTP(w, r)
		})
	}
}

// 여러 역할 중 하나를 가진 사용자만 접근을 허용
func (m *JWTMiddleware) RequireAnyRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 사용자 정보 가져옴
			userInfo, ok := r.Context().Value(UserContextKey).(*UserInfo)
			if !ok {
				m.writeForbiddenResponse(w, "사용자 정보를 찾을 수 없습니다.")
				return
			}

			// 역할 확인
			hasRole := false
			for _, role := range requiredRoles {
				if userInfo.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.writeForbiddenResponse(w, fmt.Sprintf("다음 역할 중 하나가 필요합니다: %s", strings.Join(requiredRoles, ",")))
				return
			}

			// 다음 핸들러 호출
			next.ServeHTTP(w, r)
		})
	}
}

// 선택적 인증을 제공
// 토큰이 있으면 검증하고, 없어도 요청을 계속 진행
func (m *JWTMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Authorization 헤더에서 토큰 추출 시도
		token, err := m.extractToken(r)
		if err != nil {
			// 토큰이 없으면 익명 사용자로 처리
			next.ServeHTTP(w, r)
			return
		}

		// JWT 토큰 검증 시도
		claims, err := m.jwtAuth.ValidateAccessToken(token)
		if err != nil {
			// 토큰이 유효하지 않으면 익명 사용자로 처리
			next.ServeHTTP(w, r)
			return
		}

		// 사용자 정보를 컨텍스트에 추가
		userInfo := &UserInfo{
			UserID:   claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
			Token:    token,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
		ctx = context.WithValue(ctx, TokenContextKey, token)

		// 다음 핸들러 호출
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *JWTMiddleware) extractToken(r *http.Request) (string, error) {
	// Authorization 헤더 확인
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization 헤더가 없습니다.")
	}

	// Bearer 토큰 형식 확인
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("잘못된 Authorization 헤더 형식입니다.")
	}

	token := parts[1]
	if token == "" {
		return "", fmt.Errorf("토큰이 비어있습니다.")
	}

	return token, nil
}

// 401 UnauthorizedResponse
func (m *JWTMiddleware) writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	jsonResponse := fmt.Sprintf(`{"success":false, "error":"Unauthorized", "message":"%s", "code":"AUTH_REQUIRED"}`, message)
	w.Write([]byte(jsonResponse))
}

// 403 Forbidden
func (m *JWTMiddleware) writeForbiddenResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	jsonResponse := fmt.Sprintf(`{"success":false, "error":"Forbidden","message":"%s","code":"INSUFFICIENT_PERMISSIONS"}`, message)
	w.Write([]byte(jsonResponse))
}

// 사용자 정보
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	userInfo, ok := ctx.Value(UserContextKey).(*UserInfo)
	return userInfo, ok
}

// 컨텍스트에서 토큰을 가져옴
func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenContextKey).(string)
	return token, ok
}

// 인증이 필요한 핸들러를 위한 헬퍼 함수
func RequireAuth(jwtAuth *auth.JWTAuth) func(http.Handler) http.Handler {
	middleware := NewJWTMiddleware(jwtAuth)
	return middleware.Authenticate
}

// 특정 역할이 필요한 핸들러를 위한 헬퍼 함수
func RequireRole(jwtAuth *auth.JWTAuth, role string) func(http.Handler) http.Handler {
	middleware := NewJWTMiddleware(jwtAuth)
	return middleware.RequireRole(role)
}

// 여러 역할 중 하나가 필요한 핸들러를 위한 헬퍼 함수
func RequireAnyRole(jwtAuth *auth.JWTAuth, roles ...string) func(http.Handler) http.Handler {
	middleware := NewJWTMiddleware(jwtAuth)
	return middleware.RequireAnyRole(roles...)
}

// 선택적 인증을 위한 헬퍼 함수
func OptionalAuth(jwtAuth *auth.JWTAuth) func(http.Handler) http.Handler {
	middleware := NewJWTMiddleware(jwtAuth)
	return middleware.OptionalAuth
}
