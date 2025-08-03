package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// responseWriter는 응답을 가로채기 위한 wrapper
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

// WriteHeader는 상태 코드를 기록
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write는 응답 본문을 기록
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// HTTP 요청과 응답을 로깅하는 미들웨어
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 요청 본문 읽기 (필요한 경우)
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 응답을 가로채기 위한 wrapper
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // 기본값
		}

		// 다음 핸들러 호출
		next.ServeHTTP(wrapped, r)

		// 요청 처리 시간 계산
		duration := time.Since(start)

		// 로그 메시지 생성
		logMessage := map[string]interface{}{
			"timestamp":      start.Format(time.RFC3339),
			"method":         r.Method,
			"path":           r.URL.Path,
			"query":          r.URL.RawQuery,
			"status_code":    wrapped.statusCode,
			"duration_ms":    duration.Milliseconds(),
			"user_agent":     r.UserAgent(),
			"remote_addr":    r.RemoteAddr,
			"content_type":   r.Header.Get("Content-Type"),
			"content_length": r.ContentLength,
		}

		// 요청 본문이 있는 경우 로그에 추가 (JSON인 경우만)
		if len(requestBody) > 0 && r.Header.Get("Content-Type") == "application/json" {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, requestBody, "", "  "); err == nil {
				logMessage["request_body"] = prettyJSON.String()
			}
		}

		// 응답 본문이 있는 경우 로그에 추가 (JSON인 경우만)
		if wrapped.body.Len() > 0 && wrapped.Header().Get("Content-Type") == "application/json" {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, wrapped.body.Bytes(), "", "  "); err == nil {
				logMessage["response_body"] = prettyJSON.String()
			}
		}

		// JSON 형태로 로그 출력
		logJSON, _ := json.Marshal(logMessage)
		log.Printf("HTTP Request: %s", string(logJSON))
	})
}

// 간단한 HTTP 요청 로깅을 제공하는 미들웨어
func SimpleLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 응답을 가로채기 위한 wrapper
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 다음 핸들러 호출
		next.ServeHTTP(wrapped, r)

		// 요청 처리 시간 계산
		duration := time.Since(start)

		// 간단한 로그 출력
		log.Printf("[HTTP] %s %s - %d - %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}
