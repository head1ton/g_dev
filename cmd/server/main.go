package main

import (
	"context"
	"fmt"
	"g_dev/internal/auth"
	"g_dev/internal/config"
	"g_dev/internal/database"
	"g_dev/internal/handler"
	"g_dev/internal/middleware"
	"g_dev/internal/model"
	"g_dev/internal/service"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"

	// Swagger 문서를 위한 import (자동 생성됨)
	_ "g_dev/docs"
)

// @title G-Step 웹게임서버 API
// @version 1.0
// @description G-Step 웹게임서버의 REST API 문서입니다.
// @description 이 API는 계산기 기능과 파일 처리 기능을 제공합니다.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support Team
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /

// @tag.name Calculator
// @tag.description 계산기 관련 API 엔드포인트

// @tag.name FileProcessor
// @tag.description 파일 처리 관련 API 엔드포인트

func main() {
	fmt.Println("G-Dev 게임서버를 시작합니다.")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	// 데이터베이스 연결
	dbConfig := database.NewDatabaseConfig()
	db := database.NewDatabase(dbConfig)

	if err := db.Connect(); err != nil {
		log.Fatalf("데이터베이스 연결 실패: %v", err)
	}
	defer db.Disconnect()

	// 데이터베이스 마이그레이션
	if err := db.Migrate(&model.User{}, &model.Game{}, &model.Score{}); err != nil {
		log.Fatalf("데이터베이스 마이그레이션 실패: %v", err)
	}

	// Redis 클라이언트 생성
	redisClient := createRedisClient(cfg)
	defer redisClient.Close()

	// JWT 인증 시스템 초기화
	jwtConfig := auth.NewJWTConfig(cfg)
	jwtAuth, err := auth.NewJWTAuth(jwtConfig, redisClient)
	if err != nil {
		log.Fatalf("JWT 인증 시스템 초기화 실패: %v", err)
	}

	// 서비스 레이어 초기화
	userService := service.NewUserService(db.GetDB())

	// API 핸들러 초기화
	apiHandler := handler.NewAPIHandler()
	authHandler := handler.NewAuthHandler(userService, jwtAuth)

	// 서버 포트 설정
	port := getPort()

	// HTTP 라우터 설정
	setupRoutes(apiHandler, authHandler, jwtAuth)

	fmt.Printf("서버가 http://localhost:%s 에서 실행 중입니다.\n", port)
	fmt.Printf("Swagger 문서 : http://localhost:%s/swagger/index.html\n", port)
	fmt.Printf("인증 시스템이 활성화됨")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Redis 클라이언트 생성
func createRedisClient(cfg *config.Config) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})

	// 연결 테스트
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Printf("Redis 연결 경고: %v (서버는 계속 실행됩니다)", err)
	} else {
		log.Println("Redis 연결 성공")
	}

	return redisClient
}

// HTTP 라우터 설정
func setupRoutes(apiHandler *handler.APIHandler, authHandler *handler.AuthHandler, jwtAuth *auth.JWTAuth) {
	// 정적 파일 서빙 (Swagger UI)
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"), // Swagger JSON 파일 경로
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// 홈페이지
	http.HandleFunc("/", homeHandler)

	// 인증 API
	http.HandleFunc("/api/auth/register", authHandler.HandleRegister)
	http.HandleFunc("/api/auth/login", authHandler.HandleLogin)
	http.HandleFunc("/api/auth/refresh", authHandler.HandleRefreshToken)

	// 인증이 필요한 API
	http.Handle("/api/auth/logout", middleware.RequireAuth(jwtAuth)(http.HandlerFunc(authHandler.HandleLogout)))
	http.Handle("/api/auth/profile", middleware.RequireAuth(jwtAuth)(http.HandlerFunc(authHandler.HandleProfile)))

	// 계산기 API 엔드포인트
	http.HandleFunc("/api/calculator/calculate", apiHandler.HandleCalculatorCalculate)
	http.HandleFunc("/api/calculator/history", apiHandler.HandleCalculatorHistory)
	http.HandleFunc("/api/calculator/stats", apiHandler.HandleCalculatorStats)

	// 파일 처리 API
	http.HandleFunc("/api/files/list", apiHandler.HandleFileList)
	http.HandleFunc("/api/files/search", apiHandler.HandleFileSearch)
	http.HandleFunc("/api/files/read", apiHandler.HandleFileRead)
	http.HandleFunc("/api/files/write", apiHandler.HandleFileWrite)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>G-Step 웹게임서버</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .api-section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .api-section h2 { color: #666; margin-top: 0; }
        .api-link { display: inline-block; margin: 10px 10px 10px 0; padding: 10px 15px; background: #007bff; color: white; text-decoration: none; border-radius: 5px; }
        .api-link:hover { background: #0056b3; }
        .swagger-link { background: #28a745; }
        .swagger-link:hover { background: #1e7e34; }
        .status { text-align: center; color: #28a745; font-weight: bold; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎮 G-Step 웹게임서버</h1>
        <div class="status">✅ 서버가 정상적으로 실행 중입니다!</div>
        
        <div class="api-section">
            <h2>📚 API 문서</h2>
            <a href="/swagger/index.html" class="api-link swagger-link">📖 Swagger UI</a>
        </div>

        <div class="api-section">
            <h2>🔐 인증 API (공개)</h2>
            <p>POST /api/auth/register - 회원가입</p>
            <p>POST /api/auth/login - 로그인</p>
            <p>POST /api/auth/refresh - 토큰 갱신</p>
        </div>

        <div class="api-section">
            <h2>👤 사용자 API (인증 필요)</h2>
            <p>POST /api/auth/logout - 로그아웃</p>
            <p>GET /api/auth/profile - 프로필 조회</p>
        </div>

        <div class="api-section">
            <h2>🧮 계산기 API (인증 필요)</h2>
            <p>POST /api/calculator/calculate - 계산 수행</p>
            <p>GET /api/calculator/history - 계산 히스토리</p>
            <p>GET /api/calculator/stats - 계산 통계</p>
        </div>

        <div class="api-section">
            <h2>📁 파일 처리 API (인증 필요)</h2>
            <p>POST /api/files/list - 파일 목록 조회</p>
            <p>POST /api/files/search - 파일 검색</p>
            <p>POST /api/files/read - 파일 읽기</p>
            <p>POST /api/files/write - 파일 쓰기</p>
        </div>

        <div class="api-section">
            <h2>🔧 개발 정보</h2>
            <p><strong>서버 주소:</strong> http://localhost:8080</p>
            <p><strong>API 문서:</strong> http://localhost:8080/swagger/index.html</p>
            <p><strong>프로젝트:</strong> G-Step 웹게임서버 (Go 언어)</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return port
}
