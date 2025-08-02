package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"g_dev/internal/auth"
	"g_dev/internal/config"
	"g_dev/internal/database"
	"g_dev/internal/middleware"
	"g_dev/internal/model"
	"g_dev/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// 테스트용 인증 핸들러 생성
func setupTestAuthHandler(t *testing.T) (*AuthHandler, func()) {
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-2024")
	os.Setenv("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRY", "168h")
	os.Setenv("DATABASE_HOST", "127.0.0.1")
	os.Setenv("DATABASE_PORT", "3306")
	os.Setenv("DATABASE_USERNAME", "root")
	os.Setenv("DATABASE_PASSWORD", "qwer1234!")
	os.Setenv("DATABASE_NAME", "g_dev_test")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// 데이터베이스 연결
	dbConfig := database.NewDatabaseConfig()
	db := database.NewDatabase(dbConfig)
	if err != nil {
		t.Skip("Database not available, skipping database-dependent tests")
	}

	err = db.Connect()
	if err != nil {
		t.Skip("Database connection failed, skipping database-dependent tests")
	}

	// 테이블 마이그레이션
	err = db.Migrate(&model.User{}, &model.Game{}, &model.Score{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Redis 클라이언트 생성
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15,
	})

	// Redis 연결 테스트
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis server not available, skipping Redis-dependent tests")
	}

	// 테스트 DB 초기화
	redisClient.FlushDB(ctx)

	// JWT 인증 시스템 생성
	jwtConfig := auth.NewJWTConfig(cfg)
	jwtAuth, err := auth.NewJWTAuth(jwtConfig, redisClient)
	if err != nil {
		t.Fatalf("failed to create JWT auth: %v", err)
	}

	// 사용자 서비스 생성
	userService := service.NewUserService(db.GetDB())

	// 인증 핸들러 생성
	authHandler := NewAuthHandler(userService, jwtAuth)

	// 정리 함수
	cleanup := func() {
		// Redis 데이터 초기화
		redisClient.FlushDB(ctx)
		redisClient.Close()

		// 테이블 초기화 (모든 데이터 삭제)
		gormDB := db.GetDB()
		gormDB.Exec("DELETE FROM scores") // users, games를 참조
		gormDB.Exec("DELETE FROM games")  // users를 참조할 수 있음
		gormDB.Exec("DELETE FROM users")  // 마지막에 삭제

		db.Disconnect()
	}

	return authHandler, cleanup
}

// AuthHandler 생성자 테스트
func TestNewAuthHandler(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	assert.NotNil(t, authHandler)
	assert.NotNil(t, authHandler.userService)
	assert.NotNil(t, authHandler.jwtAuth)
}

// 회원가입 API 테스트
func TestAuthHandler_HandleRegister(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	tests := []struct {
		name            string
		request         RegisterRequest
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "정상적인 회원가입",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "테스트유저",
			},
			expectedStatus:  http.StatusCreated,
			expectedSuccess: true,
		},
		{
			name: "빈 사용자명",
			request: RegisterRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "테스트유저",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name: "짧은 사용자명",
			request: RegisterRequest{
				Username: "ab",
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "테스트유저",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name: "빈 이메일",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "",
				Password: "password123",
				Nickname: "테스트유저",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name: "짧은 비밀번호",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
				Nickname: "테스트유저",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name: "빈 닉네임",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				Nickname: "",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			// 요청 생성
			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 기록
			w := httptest.NewRecorder()
			authHandler.HandleRegister(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedSuccess {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.NotNil(t, response.User)
				assert.Equal(t, tt.request.Username, response.User.Username)
				assert.Equal(t, tt.request.Email, response.User.Email)
				assert.Equal(t, tt.request.Nickname, response.User.Nickname)
			} else {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}

// 로그인 API 테스트
func TestAuthHandler_HandleLogin(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// 테스트 사용자 생성
	user := &model.User{
		Username:      "testuser_login",
		Email:         "test_login@example.com",
		Nickname:      "테스트유저",
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Level:         1,
		Gold:          1000,
		Diamond:       10,
		EmailVerified: true, // 테스트에서는 이메일 인증 완료 상태로 설정
	}
	user.SetPassword("password123")
	err := authHandler.userService.CreateUser(user)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		request         LoginRequest
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "정상적인 로그인",
			request: LoginRequest{
				Username: "testuser_login",
				Password: "password123",
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "잘못된 비밀번호",
			request: LoginRequest{
				Username: "testuser_login",
				Password: "wrongpassword",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
		{
			name: "존재하지 않는 사용자",
			request: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
		{
			name: "빈 사용자명",
			request: LoginRequest{
				Username: "",
				Password: "password123",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
		{
			name: "빈 비밀번호",
			request: LoginRequest{
				Username: "testuser_login",
				Password: "",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			// 요청 생성
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 기록
			w := httptest.NewRecorder()
			authHandler.HandleLogin(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedSuccess {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.NotNil(t, response.User)
				assert.Equal(t, tt.request.Username, response.User.Username)
			} else {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}

// 토큰 갱신 API 테스트
func TestAuthHandler_HandleRefreshToken(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// 테스트 사용자 생성 및 로그인
	user := &model.User{
		Username:      "testuser_refresh",
		Email:         "test_refresh@example.com",
		Nickname:      "테스트유저",
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Level:         1,
		Gold:          1000,
		Diamond:       10,
		EmailVerified: true, // 테스트에서는 이메일 인증 완료 상태로 설정
	}
	user.SetPassword("password123")
	err := authHandler.userService.CreateUser(user)
	assert.NoError(t, err)

	// 로그인하여 토큰 생성
	loginReq := LoginRequest{
		Username: "testuser_refresh",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(loginBody))
	loginReq2.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	authHandler.HandleLogin(loginW, loginReq2)

	var loginResponse AuthResponse
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	refreshToken := loginResponse.RefreshToken

	tests := []struct {
		name            string
		request         RefreshTokenRequest
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name: "정상적인 토큰 갱신",
			request: RefreshTokenRequest{
				RefreshToken: refreshToken,
			},
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name: "잘못된 리프레시 토큰",
			request: RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
		{
			name: "빈 리프레시 토큰",
			request: RefreshTokenRequest{
				RefreshToken: "",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			// 요청 생성
			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 기록
			w := httptest.NewRecorder()
			authHandler.HandleRefreshToken(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedSuccess {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.AccessToken)
			} else {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}

// 로그아웃 API 테스트
func TestAuthHandler_HandleLogout(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// 테스트 사용자 생성 및 로그인
	user := &model.User{
		Username:      "testuser_logout",
		Email:         "test_logout@example.com",
		Nickname:      "테스트유저_로그아웃",
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Level:         1,
		Gold:          1000,
		Diamond:       10,
		EmailVerified: true,
	}
	user.SetPassword("password123")
	err := authHandler.userService.CreateUser(user)
	assert.NoError(t, err)

	// 로그인하여 토큰 생성
	loginReq := LoginRequest{
		Username: "testuser_logout",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(loginBody))
	loginReq2.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	authHandler.HandleLogin(loginW, loginReq2)

	var loginResponse AuthResponse
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	accessToken := loginResponse.AccessToken

	tests := []struct {
		name            string
		authHeader      string
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name:            "정상적인 로그아웃",
			authHeader:      "Bearer " + accessToken,
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name:            "토큰 없음",
			authHeader:      "",
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
		{
			name:            "잘못된 토큰",
			authHeader:      "Bearer invalid-token",
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 생성
			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", bytes.NewBuffer(loginBody))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 정상적인 로그아웃 테스트의 경우 컨텍스트에 사용자 정보와 토큰 추가
			if tt.name == "정상적인 로그아웃" {
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, &middleware.UserInfo{
					UserID:   user.ID,
					Username: user.Username,
					Role:     string(user.Role),
					Token:    accessToken,
				})
				ctx = context.WithValue(ctx, middleware.TokenContextKey, accessToken)
				req = req.WithContext(ctx)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			authHandler.HandleLogout(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedSuccess {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotEmpty(t, response.Message)
			} else {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}

// 프로필 조회 API 테스트
func TestAuthHandler_HandleProfile(t *testing.T) {
	authHandler, cleanup := setupTestAuthHandler(t)
	defer cleanup()

	// 테스트 사용자 생성 및 로그인
	user := &model.User{
		Username:      "testuser_profile",
		Email:         "test_profile@example.com",
		Nickname:      "테스트유저_프로필",
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Level:         1,
		Gold:          1000,
		Diamond:       10,
		EmailVerified: true, // 테스트에서는 이메일 인증 완료 상태로 설정
	}
	user.SetPassword("password123")
	err := authHandler.userService.CreateUser(user)
	assert.NoError(t, err)

	// 로그인하여 토큰 생성
	loginReq := LoginRequest{
		Username: "testuser_profile",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(loginBody))
	loginReq2.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	authHandler.HandleLogin(loginW, loginReq2)

	var loginResponse AuthResponse
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	accessToken := loginResponse.AccessToken

	tests := []struct {
		name            string
		authHeader      string
		expectedStatus  int
		expectedSuccess bool
	}{
		{
			name:            "정상적인 프로필 조회",
			authHeader:      "Bearer " + accessToken,
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
		},
		{
			name:            "토큰 없음",
			authHeader:      "",
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
		{
			name:            "잘못된 토큰",
			authHeader:      "Bearer invalid-token",
			expectedStatus:  http.StatusUnauthorized,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 생성
			req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 정상적인 프로필 조회 테스트의 경우 컨텍스트에 사용자 정보와 토큰 추가
			if tt.name == "정상적인 프로필 조회" {
				ctx := context.WithValue(req.Context(), middleware.UserContextKey, &middleware.UserInfo{
					UserID:   user.ID,
					Username: user.Username,
					Role:     string(user.Role),
					Token:    accessToken,
				})
				ctx = context.WithValue(ctx, middleware.TokenContextKey, accessToken)
				req = req.WithContext(ctx)
			}

			// 응답 기록
			w := httptest.NewRecorder()
			authHandler.HandleProfile(w, req)

			// 결과 확인
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedSuccess {
				var response AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.User)
				assert.Equal(t, "testuser_profile", response.User.Username)
				assert.Equal(t, "test_profile@example.com", response.User.Email)
				assert.Equal(t, "테스트유저_프로필", response.User.Nickname)
			} else {
				var response APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.NotEmpty(t, response.Error)
			}
		})
	}
}
