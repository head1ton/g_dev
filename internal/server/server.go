package server

import (
	"context"
	"fmt"
	"g_dev/internal/auth"
	"g_dev/internal/config"
	"g_dev/internal/database"
	"g_dev/internal/handler"
	"g_dev/internal/migration"
	"g_dev/internal/router"
	"g_dev/internal/service"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 메인 구조체
type Server struct {
	Config           *config.Config
	DB               *database.Database
	MigrationManager *migration.MigrationManager
	RedisClient      *redis.Client
	JWTAuth          *auth.JWTAuth
	UserService      *service.UserService
	APIHandler       *handler.APIHandler
	AuthHandler      *handler.AuthHandler
	Router           *router.Router
	HTTPServer       *http.Server
	Port             string
}

// 새로운 Server 인스턴스 생성
func NewServer() (*Server, error) {
	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("설정 로드 실패: %v", err)
	}

	server := &Server{
		Config: cfg,
		Port:   getPort(),
	}

	return server, nil
}

// 서버의 모든 컴포넌트를 초기화
func (s *Server) Initialize() error {
	log.Println("서버 컴포넌트 초기화 시작...")

	// 1. 데이터베이스 초기화
	if err := s.initializeDatabase(); err != nil {
		return fmt.Errorf("데이터베이스 초기화 실패: %v", err)
	}

	// 2. Redis 초기화
	if err := s.initializeRedis(); err != nil {
		return fmt.Errorf("Redis 초기화 실패: %v", err)
	}

	// 3. JWT 인증 시스템 초기화
	if err := s.initializeJWT(); err != nil {
		return fmt.Errorf("JWT 인증 시스템 초기화 실패: %v", err)
	}

	// 4. 서비스 레이어 초기화
	s.initializeServices()

	// 5. 핸들러 초기화
	s.initializeHandlers()

	// 6. 라우터 초기화
	s.initializeRouter()

	// 7. HTTP 서버 초기화
	s.initializeHTTPServer()

	log.Println("서버 컴포넌트 초기화 완료")
	return nil
}

// 데이터베이스 초기화
func (s *Server) initializeDatabase() error {
	log.Println("데이터베이스 초기화 중...")

	// 데이터베이스 연결
	dbConfig := database.NewDatabaseConfig()
	s.DB = database.NewDatabase(dbConfig)

	if err := s.DB.Connect(); err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %v", err)
	}

	// 마이그레이션 매니저 초기화
	s.MigrationManager = migration.NewMigrationManager(s.DB)
	s.MigrationManager.RegisterDefaultModels()

	// 마이그레이션 실행
	if err := s.MigrationManager.Migrate(); err != nil {
		return fmt.Errorf("데이터베이스 마이그레이션 실패: %v", err)
	}

	log.Println("데이터베이스 초기화 완료")
	return nil
}

// Redis 초기화
func (s *Server) initializeRedis() error {
	log.Println("Redis 초기화 중...")

	s.RedisClient = redis.NewClient(&redis.Options{
		Addr:     s.Config.Redis.Host + ":" + s.Config.Redis.Port,
		Password: s.Config.Redis.Password,
		DB:       0,
	})

	// Redis 연결 테스트
	ctx := context.Background()
	if err := s.RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 연결 실패: %v", err)
	}

	log.Println("Redis 초기화 완료")
	return nil
}

// JWT 인증 시스템 초기화
func (s *Server) initializeJWT() error {
	log.Println("JWT 인증 시스템 초기화 중...")

	jwtConfig := auth.NewJWTConfig(s.Config)
	jwtAuth, err := auth.NewJWTAuth(jwtConfig, s.RedisClient)
	if err != nil {
		return fmt.Errorf("JWT 인증 시스템 초기화 실패: %v", err)
	}

	s.JWTAuth = jwtAuth
	log.Println("JWT 인증 시스템 초기화 완료")
	return nil
}

// 서비스 레이어 초기화
func (s *Server) initializeServices() {
	log.Println("서비스 레이어 초기화 중...")

	s.UserService = service.NewUserService(s.DB.GetDB())

	log.Println("서비스 레이어 초기화 완료")
}

// 핸들러 초기화
func (s *Server) initializeHandlers() {
	log.Println("핸들러 초기화 중...")

	s.APIHandler = handler.NewAPIHandler()
	s.AuthHandler = handler.NewAuthHandler(s.UserService, s.JWTAuth)

	log.Println("핸들러 초기화 완료")
}

// 라우터 초기화
func (s *Server) initializeRouter() {
	log.Println("라우터 초기화 중...")

	s.Router = router.NewRouter(s.APIHandler, s.AuthHandler, s.JWTAuth, s.Port)
	s.Router.SetupRoutes()

	log.Println("라우터 초기화 완료")
}

// HTTP 서버 초기화
func (s *Server) initializeHTTPServer() {
	log.Println("HTTP 서버 초기화 중...")

	s.HTTPServer = &http.Server{
		Addr:         ":" + s.Port,
		Handler:      nil, // http.DefaultServeMux 사용
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("HTTP 서버 초기화 완료")
}

// 서버 시작
func (s *Server) Start() error {
	log.Printf("서버가 http://localhost:%s 에서 실행 중입니다....", s.Port)
	log.Printf("Swagger 문서: http://localhost:%s/swagger/index.html", s.Port)
	log.Printf("인증 시스템이 활성화되었습니다")

	// 서버 시작
	go func() {
		if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("서버 시작 실패: %v", err)
		}
	}()

	// 종료 신호 대기
	s.waitForShutdown()

	return nil
}

// 서버 종료 신호 대기
func (s *Server) waitForShutdown() {
	// 종료 신호 채널 생성
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 종료 신호 대기
	<-quit
	log.Println("서버 종료 신호를 받았음...")

	// 정상 종료를 위한 컨텍스트 생성 (30초 타임아웃)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 서버 종료
	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		log.Printf("서버 종료 중 오류: %v", err)
	}

	// 리소스 정리
	s.cleanup()

	log.Println("서버가 정상적으로 종료되었습니다")
}

// 서버 리소스 정리
func (s *Server) cleanup() {
	log.Println("서버 리소스 정리 중...")

	if s.RedisClient != nil {
		s.RedisClient.Close()
	}

	if s.DB != nil {
		s.DB.Disconnect()
	}

	log.Println("서버 리소스 정리 완료")
}

// 서버 포트 결정
func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8081"
}
