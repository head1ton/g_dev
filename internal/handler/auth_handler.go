package handler

import (
	"encoding/json"
	"fmt"
	"g_dev/internal/auth"
	"g_dev/internal/middleware"
	"g_dev/internal/model"
	"g_dev/internal/service"
	"net/http"
)

// 인증 관련 API 처리 핸들러
type AuthHandler struct {
	userService *service.UserService
	jwtAuth     *auth.JWTAuth
}

// 새로운 AuthHandler 인스턴스를 생성
func NewAuthHandler(userService *service.UserService, jwtAuth *auth.JWTAuth) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtAuth:     jwtAuth,
	}
}

// 회원가입 요청을 담는 구조체
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=50"`
	Nickname string `json:"nickname" validate:"required,min=2,max=30"`
}

// 로그인 요청을 담는 구조체
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// 토큰 갱신 요청
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// 사용자 정보
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Level    int    `json:"level"`
	Gold     int    `json:"gold"`
	Diamond  int    `json:"diamond"`
}

// 인증 응답
type AuthResponse struct {
	Success      bool      `json:"success"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	User         *UserInfo `json:"user,omitempty"`
	Message      string    `json:"message,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// 회원가입 API를 처리.
// @Summary 회원가입
// @Description 새로운 사용자를 등록.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "회원가입 정보"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} APIResponse
// @Failure 409 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 요청 본문 파싱
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return
	}

	// 요청 검증
	if err := validateRegisterRequest(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 사용자 생성
	user := &model.User{
		Username:      req.Username,
		Email:         req.Email,
		Nickname:      req.Nickname,
		Status:        model.UserStatusActive,
		Role:          model.UserRoleUser,
		Level:         1,
		Gold:          1000,
		Diamond:       10,
		EmailVerified: true, // 개발 환경에서는 이메일 인증 패스~
	}

	// 비밀번호 설정
	if err := user.SetPassword(req.Password); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "비밀번호 설정 중 오류가 발생했습니다")
		return
	}

	// 사용자 저장
	if err := h.userService.CreateUser(user); err != nil {
		if err.Error() == "username already exists" {
			writeErrorResponse(w, http.StatusConflict, "이미 사용 중인 사용자명")
			return
		}
		if err.Error() == "email already exists" {
			writeErrorResponse(w, http.StatusConflict, "이미 사용 중인 이메일")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "사용자 생성 중 오류가 발생했습니다")
		return
	}

	// JWT 토큰 생성
	accessToken, refreshToken, err := h.jwtAuth.GenerateTokenPair(user.ID, user.Username, string(user.Role))
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "토큰 생성 중 오류가 발생했습니다")
		return
	}

	// 응답 생성
	response := AuthResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Role:     string(user.Role),
			Level:    user.Level,
			Gold:     user.Gold,
			Diamond:  user.Diamond,
		},
		Message: "회원가입이 완료되었습니다",
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

// 로그인 API를 처리.
// @Summary 로그인
// @Description 사용자 로그인을 처리.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "로그인 정보"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 요청 본문 파싱
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식")
		return
	}

	// 요청 검증
	if err := validateLoginRequest(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 사용자 인증
	user, err := h.userService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "사용자명 또는 비밀번호가 올바르지 않습니다")
		return
	}

	// 계정 상태 확인
	if !user.CanLogin() {
		writeErrorResponse(w, http.StatusUnauthorized, "로그인 할 수 없는 계정")
		return
	}

	// 마지막 로그인 시간 업데이트
	user.UpdateLastLogin("")

	// JWT 토큰 생성
	accessToken, refreshToken, err := h.jwtAuth.GenerateTokenPair(user.ID, user.Username, string(user.Role))
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "토큰 생성 중 오류가 발생했습니다")
		return
	}

	// 응답 생성
	response := AuthResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Role:     string(user.Role),
			Level:    user.Level,
			Gold:     user.Gold,
			Diamond:  user.Diamond,
		},
		Message: "로그인이 완료되었습니다",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// 토큰 갱신 API를 처리
// @Summary 토큰 갱신
// @Description 리프레시 토큰을 사용하여 새로운 액세스 토큰을 발급.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "리프레시 토큰"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 요청 본문 파싱
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "잘못된 요청 형식입니다")
		return
	}

	// 요청 검증
	if err := validateRefreshTokenRequest(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 액세스 토큰 갱신
	accessToken, err := h.jwtAuth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "유효하지 않은 리프레시 토큰입니다")
		return
	}

	// 응답 생성
	response := AuthResponse{
		Success:     true,
		AccessToken: accessToken,
		Message:     "토큰이 갱신되었습니다",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// 로그아웃 API를 처리
// @Summary 로그아웃
// @Description 사용자 로그아웃을 처리합니다.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 401 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 사용자 정보 가져오기 (미들웨어에서 설정)
	userInfo, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "인증이 필요합니다")
		return
	}

	// 토큰 가져오기
	token, ok := middleware.GetTokenFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "토큰을 찾을 수 없습니다")
		return
	}

	// 로그아웃 처리
	if err := h.jwtAuth.Logout(token); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "로그아웃 처리 중 오류가 발생했습니다")
		return
	}

	// 응답 생성
	response := APIResponse{
		Success: true,
		Message: fmt.Sprintf("사용자 %s가 로그아웃되었습니다", userInfo.Username),
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// 사용자 프로필 조회 API를 처리합니다.
// @Summary 프로필 조회
// @Description 현재 로그인한 사용자의 프로필 정보를 조회합니다.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthResponse
// @Failure 401 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Router /api/auth/profile [get]
func (h *AuthHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 사용자 정보 가져오기 (미들웨어에서 설정)
	userInfo, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "인증이 필요합니다")
		return
	}

	// 데이터베이스에서 최신 사용자 정보 조회
	user, err := h.userService.GetUserByID(userInfo.UserID)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "사용자를 찾을 수 없습니다")
		return
	}

	// 응답 생성
	response := AuthResponse{
		Success: true,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Role:     string(user.Role),
			Level:    user.Level,
			Gold:     user.Gold,
			Diamond:  user.Diamond,
		},
		Message: "프로필 정보를 조회했습니다",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func validateRefreshTokenRequest(req *RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return fmt.Errorf("리프레시 토큰은 필수입니다")
	}
	return nil
}

func validateLoginRequest(req *LoginRequest) error {
	if req.Username == "" {
		return fmt.Errorf("사용자명은 필수")
	}
	if req.Password == "" {
		return fmt.Errorf("비밀번호는 필수")
	}
	return nil
}

func validateRegisterRequest(req *RegisterRequest) error {
	if req.Username == "" {
		return fmt.Errorf("사용자명은 필수")
	}
	if len(req.Username) < 3 || len(req.Username) > 20 {
		return fmt.Errorf("사용자명은 3-20자 사이여야 ")
	}
	if req.Email == "" {
		return fmt.Errorf("이메일은 필수")
	}
	if req.Password == "" {
		return fmt.Errorf("비밀번호는 필수")
	}
	if len(req.Password) < 6 || len(req.Password) > 50 {
		return fmt.Errorf("비밀번호는 6-50자 사이여야 ")
	}
	if req.Nickname == "" {
		return fmt.Errorf("닉네임은 필수")
	}
	if len(req.Nickname) < 2 || len(req.Nickname) > 30 {
		return fmt.Errorf("닉네임은 2-30자 사이여야 ")
	}
	return nil
}

// 에러 응답
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}

	writeJSONResponse(w, statusCode, response)
}

// JSON 응답
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
