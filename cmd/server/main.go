package main

import (
	"fmt"
	"g_dev/internal/handler"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"

	// Swagger 문서를 위한 import (자동 생성됨)
	_ "g_dev/docs"
)

func main() {
	fmt.Println("G-Dev 게임서버를 시작합니다.")

	port := getPort()

	// API 핸들러 초기화
	apiHandler := handler.NewAPIHandler()

	// HTTP 라우터 설정
	setupRoutes(apiHandler)

	fmt.Printf("서버가 http://localhost:%s 에서 실행 중입니다.\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// HTTP 라우터 설정
func setupRoutes(apiHandler *handler.APIHandler) {
	// 정적 파일 서빙 (Swagger UI)
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"), // Swagger JSON 파일 경로
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// 홈페이지
	http.HandleFunc("/", homeHandler)

	// 계산기 API 엔드포인트
	http.HandleFunc("/api/calculator/calculate", apiHandler.HandleCalculatorCalculate)
	http.HandleFunc("/api/calculator/history", apiHandler.HandleCalculatorHistory)
	http.HandleFunc("/api/calculator/stats", apiHandler.HandleCalculatorStats)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	html := `
<!DOCTYPE html>
<html>
<head>
    <title>G-Step 웹게임서버</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            text-align: center; 
            margin-top: 50px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .container {
            background: rgba(255,255,255,0.1);
            padding: 30px;
            border-radius: 15px;
            backdrop-filter: blur(10px);
            max-width: 800px;
            margin: 0 auto;
        }
        h1 { color: #fff; }
        .status { 
            background: rgba(0,255,0,0.2); 
            padding: 10px; 
            border-radius: 5px;
            margin: 20px 0;
        }
        .api-links {
            display: flex;
            justify-content: center;
            gap: 20px;
            margin: 20px 0;
        }
        .api-link {
            background: rgba(255,255,255,0.2);
            padding: 15px 25px;
            border-radius: 8px;
            text-decoration: none;
            color: white;
            transition: all 0.3s ease;
        }
        .api-link:hover {
            background: rgba(255,255,255,0.3);
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎮 G-Step 웹게임서버</h1>
        <div class="status">
            <h2>✅ 서버가 정상적으로 실행 중입니다!</h2>
            <p>Go 언어로 개발된 웹게임서버입니다.</p>
        </div>
        
        <div class="api-links">
            <a href="/swagger/index.html" class="api-link" target="_blank">
                📚 API 문서 (Swagger)
            </a>
            <a href="/api/calculator/stats" class="api-link" target="_blank">
                📊 계산기 통계
            </a>
        </div>
        
        <p>현재 시간: <span id="time"></span></p>
    </div>
    <script>
        function updateTime() {
            document.getElementById('time').textContent = new Date().toLocaleString('ko-KR');
        }
        updateTime();
        setInterval(updateTime, 1000);
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return port
}
