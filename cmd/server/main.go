package main

import (
	"g_dev/internal/server"
	"log"
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

// @tag.name Auth
// @tag.description 인증 관련 API 엔드포인트

func main() {
	log.Println("G-Dev 게임서버를 시작합니다.")

	// 서버 인스턴스 생성
	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("서버 생성 실패: %v", err)
	}

	// 서버 초기화
	if err := srv.Initialize(); err != nil {
		log.Fatalf("서버 초기화 실패: %v", err)
	}

	// 서버 시작
	if err := srv.Start(); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
