package router

import (
	"g_dev/internal/auth"
	"g_dev/internal/handler"
	"g_dev/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

// HTTP 라우터 설정
type Router struct {
	// 핸들러들
	APIHandler  *handler.APIHandler
	AuthHandler *handler.AuthHandler

	// 인증 시스템
	JWTAuth *auth.JWTAuth

	// 서버 설정
	Port string
}

// 새로운 Router 인스턴스 생성
func NewRouter(apiHandler *handler.APIHandler, authHandler *handler.AuthHandler, jwtAuth *auth.JWTAuth, port string) *Router {
	return &Router{
		APIHandler:  apiHandler,
		AuthHandler: authHandler,
		JWTAuth:     jwtAuth,
		Port:        port,
	}
}

// 모든 HTTP 라우트를 설정
func (r *Router) SetupRoutes() {
	// Swagger 문서 라우트
	r.setupSwaggerRoutes()

	// 정적 파일 및 홈페이지 라우트
	r.setupStaticsRoutes()

	// 인증 API 라우트
	r.setupPublicAuthRoutes()

	// 보호된 API 라우트
	r.setupProtectedRoutes()
}

// Swagger 문서 라우트 설정
func (r *Router) setupSwaggerRoutes() {
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+r.Port+"/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}

// 정적 파일과 홈페이지 라우트 설정
func (r *Router) setupStaticsRoutes() {
	http.HandleFunc("/", r.homeHandler)
}

// 공개 인증 API 라우트 설정
func (r *Router) setupPublicAuthRoutes() {
	// 회원가입
	//http.HandleFunc("/api/auth/register", r.AuthHandler.HandleRegister)
	http.Handle("/api/auth/register", middleware.SimpleLoggingMiddleware(http.HandlerFunc(r.AuthHandler.HandleRegister)))

	// 로그인
	//http.HandleFunc("/api/auth/login", r.AuthHandler.HandleLogin)
	http.Handle("/api/auth/login", middleware.SimpleLoggingMiddleware(http.HandlerFunc(r.AuthHandler.HandleLogin)))

	// 토큰 갱신
	//http.HandleFunc("/api/auth/refresh", r.AuthHandler.HandleRefreshToken)
	http.Handle("/api/uath/refresh", middleware.SimpleLoggingMiddleware(http.HandlerFunc(r.AuthHandler.HandleRefreshToken)))
}

// 인증이 필요한 보호된 API 라우트 설정
func (r *Router) setupProtectedRoutes() {
	// 인증이 필요한 API 엔드포인트들
	protectedRoutes := []struct {
		path    string
		handler http.HandlerFunc
	}{
		// 인증 관련 (보호됨)
		{"/api/auth/logout", r.AuthHandler.HandleLogout},
		{"/api/auth/profile", r.AuthHandler.HandleProfile},

		// 계산기 API (보호됨)
		{"/api/calculator/calculate", r.APIHandler.HandleCalculatorCalculate},
		{"/api/calculator/history", r.APIHandler.HandleCalculatorHistory},
		{"/api/calculator/stats", r.APIHandler.HandleCalculatorStats},

		// 파일 처리 API (보호됨)
		{"/api/files/list", r.APIHandler.HandleFileList},
		{"/api/files/search", r.APIHandler.HandleFileSearch},
		{"/api/files/read", r.APIHandler.HandleFileRead},
		{"/api/files/write", r.APIHandler.HandleFileWrite},
	}

	// 각 보호된 라우트에 JWT 인증 미들웨어 적용
	for _, route := range protectedRoutes {
		//http.Handle(route.path, middleware.RequireAuth(r.JWTAuth)(http.HandlerFunc(route.handler)))
		http.Handle(route.path, middleware.SimpleLoggingMiddleware(middleware.RequireAuth(r.JWTAuth)(http.HandlerFunc(route.handler))))
	}
}

func (r *Router) homeHandler(w http.ResponseWriter, req *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>G-Step 웹게임서버</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .section { margin: 20px 0; padding: 15px; border-left: 4px solid #007bff; background-color: #f8f9fa; }
        .endpoint { margin: 10px 0; padding: 10px; background-color: #e9ecef; border-radius: 5px; }
        .method { font-weight: bold; color: #007bff; }
        .url { font-family: monospace; color: #28a745; }
        .description { color: #666; margin-top: 5px; }
        .auth-required { color: #dc3545; font-weight: bold; }
        .public { color: #28a745; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>G-Dev 웹게임서버</h1>
        
        <div class="section">
            <h2>API 문서</h2>
            <div class="endpoint">
                <span class="method">GET</span> <span class="url">/swagger/index.html</span>
                <div class="description">Swagger API 문서</div>
            </div>
        </div>

        <div class="section">
            <h2>인증 API <span class="public">(공개)</span></h2>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/auth/register</span>
                <div class="description">회원가입</div>
            </div>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/auth/login</span>
                <div class="description">로그인</div>
            </div>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/auth/refresh</span>
                <div class="description">토큰 갱신</div>
            </div>
        </div>

        <div class="section">
            <h2>사용자 API <span class="auth-required">(인증 필요)</span></h2>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/auth/logout</span>
                <div class="description">로그아웃</div>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> <span class="url">/api/auth/profile</span>
                <div class="description">프로필 조회</div>
            </div>
        </div>

        <div class="section">
            <h2>계산기 API <span class="auth-required">(인증 필요)</span></h2>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/calculator/calculate</span>
                <div class="description">계산 수행</div>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> <span class="url">/api/calculator/history</span>
                <div class="description">계산 히스토리 조회</div>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> <span class="url">/api/calculator/stats</span>
                <div class="description">계산 통계 조회</div>
            </div>
        </div>

        <div class="section">
            <h2>파일 처리 API <span class="auth-required">(인증 필요)</span></h2>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/files/list</span>
                <div class="description">파일 목록 조회</div>
            </div>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/files/search</span>
                <div class="description">파일 검색</div>
            </div>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/files/read</span>
                <div class="description">파일 읽기</div>
            </div>
            <div class="endpoint">
                <span class="method">POST</span> <span class="url">/api/files/write</span>
                <div class="description">파일 쓰기</div>
            </div>
        </div>

        <div class="section">
            <h2>서버 정보</h2>
            <p><strong>포트:</strong> ` + r.Port + `</p>
            <p><strong>상태:</strong> <span style="color: #28a745;">실행 중</span></p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
