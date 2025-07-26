package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("G-Dev 게임서버를 시작합니다.")

	port := getPort()

	http.HandleFunc("/", homeHandler)

	fmt.Printf("서버가 http://localhost:%s 에서 실행 중입니다.\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
            max-width: 600px;
            margin: 0 auto;
        }
        h1 { color: #fff; }
        .status { 
            background: rgba(0,255,0,0.2); 
            padding: 10px; 
            border-radius: 5px;
            margin: 20px 0;
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
