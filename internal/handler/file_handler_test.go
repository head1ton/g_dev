package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHandleFileList는 파일 목록 조회 API를 테스트
// 다양한 시나리오와 에러 케이스를 포함하여 API의 안정성을 검증
func TestHandleFileList(t *testing.T) {
	handler := NewAPIHandler()

	// 테스트용 임시 디렉토리 생성 (상대 경로 사용)
	tempDir := "test_temp_dir"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 테스트 후 정리

	// 테스트용 파일들 생성
	testFiles := []string{"test1.txt", "test2.txt", "test3.txt"}
	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		content := []byte("test content for " + filename)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// 테스트용 디렉토리 생성
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		requestBody    FileListRequest
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:   "정상적인 파일 목록 조회",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path:       tempDir,
				SortBy:     "name",
				SortOrder:  "asc",
				ShowHidden: false,
				MaxResults: 100,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "기본값으로 목록 조회",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path: "", // 기본값 사용
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "잘못된 HTTP 메서드",
			method: http.MethodGet,
			requestBody: FileListRequest{
				Path: tempDir,
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			errorContains:  "Method not allowed",
		},
		{
			name:   "잘못된 정렬 기준",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path:   tempDir,
				SortBy: "invalid_sort",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid sort_by value",
		},
		{
			name:   "잘못된 정렬 순서",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path:      tempDir,
				SortOrder: "invalid_order",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid sort_order value",
		},
		{
			name:   "최대 결과 수 초과",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path:       tempDir,
				MaxResults: 2000, // 1000 초과
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "max_results cannot exceed 1000",
		},
		{
			name:   "절대 경로 사용 (보안 검사)",
			method: http.MethodPost,
			requestBody: FileListRequest{
				Path: "/etc/passwd", // 시스템 디렉토리
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "access to system directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/files/list", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 레코더 생성
			recorder := httptest.NewRecorder()

			// 핸들러 호출
			handler.HandleFileList(recorder, req)

			// 상태 코드 확인
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// 성공 케이스인 경우 응답 내용 확인
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
					return
				}

				// 응답 데이터 구조 확인
				if response.Data == nil {
					t.Error("Expected non-nil data in response")
				}
			}

			// 에러 케이스인 경우 에러 메시지 확인
			if tt.expectError && tt.errorContains != "" {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}

				if !strings.Contains(response.Error, tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, response.Error)
				}
			}
		})
	}
}

// TestHandleFileSearch는 파일 검색 API를 테스트
// 다양한 검색 조건과 에러 케이스를 검증
func TestHandleFileSearch(t *testing.T) {
	handler := NewAPIHandler()

	// 테스트용 임시 디렉토리 생성 (상대 경로 사용)
	tempDir := "test_search_dir"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 테스트 후 정리

	// 테스트용 파일들 생성
	testFiles := map[string]string{
		"test1.txt":    "Hello World",
		"test2.txt":    "Hello Go",
		"config.json":  `{"name": "test", "value": 123}`,
		"data.csv":     "name,age\nJohn,25\nJane,30",
		"script.py":    "print('Hello Python')",
		"document.pdf": "PDF content",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name           string
		method         string
		requestBody    FileSearchRequest
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:   "와일드카드 패턴으로 검색",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Pattern: "*.txt",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "확장자로 검색",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Extension: "json",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "내용으로 검색",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Content:       "Hello",
				CaseSensitive: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "정규표현식으로 검색",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				RegexPattern: ".*\\.txt$",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "검색 조건 없음",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				// 모든 검색 조건이 비어있음
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "at least one search condition must be specified",
		},
		{
			name:   "너무 광범위한 패턴",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Pattern: "*", // 너무 광범위한 패턴
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "too broad search pattern",
		},
		{
			name:   "너무 긴 정규표현식",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				RegexPattern: strings.Repeat("a", 101), // 100자 초과
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "regex pattern is too long",
		},
		{
			name:   "너무 긴 확장자",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Extension: strings.Repeat("a", 21), // 20자 초과
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "extension is too long",
		},
		{
			name:   "너무 긴 검색 내용",
			method: http.MethodPost,
			requestBody: FileSearchRequest{
				Content: strings.Repeat("a", 201), // 200자 초과
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "search content is too long",
		},
		{
			name:   "잘못된 HTTP 메서드",
			method: http.MethodGet,
			requestBody: FileSearchRequest{
				Pattern: "*.txt",
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			errorContains:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/files/search", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 레코더 생성
			recorder := httptest.NewRecorder()

			// 핸들러 호출
			handler.HandleFileSearch(recorder, req)

			// 상태 코드 확인
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// 성공 케이스인 경우 응답 내용 확인
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
					return
				}

				// 검색 결과가 배열인지 확인
				if results, ok := response.Data.([]interface{}); ok {
					t.Logf("Found %d search results", len(results))
				} else {
					t.Error("Expected search results to be an array")
				}
			}

			// 에러 케이스인 경우 에러 메시지 확인
			if tt.expectError && tt.errorContains != "" {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}

				if !strings.Contains(response.Error, tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, response.Error)
				}
			}
		})
	}
}

// TestHandleFileRead는 파일 읽기 API를 테스트
// 파일 읽기 기능과 다양한 에러 케이스를 검증
func TestHandleFileRead(t *testing.T) {
	handler := NewAPIHandler()

	// 테스트용 임시 디렉토리 생성 (상대 경로 사용)
	tempDir := "test_read_dir"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 테스트 후 정리

	// 테스트용 파일 생성
	testContent := "Hello World\nThis is a test file\nWith multiple lines"
	testFilePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 큰 파일 생성 (크기 제한 테스트용)
	largeContent := strings.Repeat("Large file content\n", 10000) // 약 200KB
	largeFilePath := filepath.Join(tempDir, "large.txt")
	if err := os.WriteFile(largeFilePath, []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		requestBody    FileReadRequest
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:   "정상적인 파일 읽기",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path:        testFilePath,
				Encoding:    "utf-8",
				MaxSize:     1024 * 1024, // 1MB
				LineNumbers: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "기본값으로 파일 읽기",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path: testFilePath,
				// 다른 필드들은 기본값 사용
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "존재하지 않는 파일",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path: filepath.Join(tempDir, "nonexistent.txt"),
			},
			expectedStatus: http.StatusNotFound, // FileProcessor에서 404 반환
			expectError:    true,
		},
		{
			name:   "파일 경로 없음",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path: "", // 빈 경로
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "file path is required",
		},
		{
			name:   "절대 경로 사용 (보안 검사)",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path: "/etc/passwd", // 시스템 디렉토리
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "access to system directory",
		},
		{
			name:   "잘못된 인코딩",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path:     testFilePath,
				Encoding: "invalid-encoding",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid encoding",
		},
		{
			name:   "파일 크기 제한 초과",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path:    largeFilePath,
				MaxSize: 1024, // 1KB로 제한
			},
			expectedStatus: http.StatusInternalServerError, // FileProcessor에서 에러 발생
			expectError:    true,
		},
		{
			name:   "최대 크기 제한 초과",
			method: http.MethodPost,
			requestBody: FileReadRequest{
				Path:    testFilePath,
				MaxSize: 20 * 1024 * 1024, // 20MB (10MB 초과)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "max_size cannot exceed 10MB",
		},
		{
			name:   "잘못된 HTTP 메서드",
			method: http.MethodGet,
			requestBody: FileReadRequest{
				Path: testFilePath,
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			errorContains:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/files/read", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 레코더 생성
			recorder := httptest.NewRecorder()

			// 핸들러 호출
			handler.HandleFileRead(recorder, req)

			// 상태 코드 확인
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// 성공 케이스인 경우 응답 내용 확인
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
					return
				}

				// 파일 읽기 응답 구조 확인
				if readData, ok := response.Data.(map[string]interface{}); ok {
					if content, exists := readData["content"]; exists {
						t.Logf("File content length: %d", len(content.(string)))
					}
					if size, exists := readData["size"]; exists {
						t.Logf("File size: %v", size)
					}
					if lineCount, exists := readData["line_count"]; exists {
						t.Logf("Line count: %v", lineCount)
					}
				} else {
					t.Error("Expected file read response to be a map")
				}
			}

			// 에러 케이스인 경우 에러 메시지 확인
			if tt.expectError && tt.errorContains != "" {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}

				if !strings.Contains(response.Error, tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, response.Error)
				}
			}
		})
	}
}

// TestHandleFileWrite는 파일 쓰기 API를 테스트
// 파일 쓰기 기능과 다양한 에러 케이스를 검증
func TestHandleFileWrite(t *testing.T) {
	handler := NewAPIHandler()

	// 테스트용 임시 디렉토리 생성 (상대 경로 사용)
	tempDir := "test_write_dir"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 테스트 후 정리

	tests := []struct {
		name           string
		method         string
		requestBody    FileWriteRequest
		expectedStatus int
		expectError    bool
		errorContains  string
	}{
		{
			name:   "새 파일 생성",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:      filepath.Join(tempDir, "newfile.txt"),
				Content:   "Hello World\nThis is a new file",
				Encoding:  "utf-8",
				Append:    false,
				CreateDir: false,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "기본값으로 파일 쓰기",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:    filepath.Join(tempDir, "default.txt"),
				Content: "Default content",
				// 다른 필드들은 기본값 사용
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "디렉토리 자동 생성",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:      filepath.Join(tempDir, "subdir", "file.txt"),
				Content:   "Content in subdirectory",
				CreateDir: true,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "파일 경로 없음",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:    "", // 빈 경로
				Content: "test content",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "file path is required",
		},
		{
			name:   "절대 경로 사용 (보안 검사)",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:    "/etc/test.txt", // 시스템 디렉토리
				Content: "test content",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "access to system directory",
		},
		{
			name:   "너무 큰 파일 내용",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:    filepath.Join(tempDir, "large.txt"),
				Content: strings.Repeat("Large content\n", 100000), // 약 2MB
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "file content is too large",
		},
		{
			name:   "잘못된 인코딩",
			method: http.MethodPost,
			requestBody: FileWriteRequest{
				Path:     filepath.Join(tempDir, "test.txt"),
				Content:  "test content",
				Encoding: "invalid-encoding",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorContains:  "invalid encoding",
		},
		{
			name:   "잘못된 HTTP 메서드",
			method: http.MethodGet,
			requestBody: FileWriteRequest{
				Path:    filepath.Join(tempDir, "test.txt"),
				Content: "test content",
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
			errorContains:  "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 요청 본문 생성
			requestBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(tt.method, "/api/files/write", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// 응답 레코더 생성
			recorder := httptest.NewRecorder()

			// 핸들러 호출
			handler.HandleFileWrite(recorder, req)

			// 상태 코드 확인
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// 성공 케이스인 경우 응답 내용 확인
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
					return
				}

				if !response.Success {
					t.Errorf("Expected success response, got error: %s", response.Error)
					return
				}

				// 파일 쓰기 응답 구조 확인
				if writeData, ok := response.Data.(map[string]interface{}); ok {
					if size, exists := writeData["size"]; exists {
						t.Logf("Written file size: %v", size)
					}
					if lineCount, exists := writeData["line_count"]; exists {
						t.Logf("Line count: %v", lineCount)
					}
					if created, exists := writeData["created"]; exists {
						t.Logf("File created: %v", created)
					}
				} else {
					t.Error("Expected file write response to be a map")
				}

				// 실제 파일이 생성되었는지 확인
				if tt.requestBody.Path != "" {
					if _, err := os.Stat(tt.requestBody.Path); os.IsNotExist(err) {
						t.Errorf("Expected file to be created at %s", tt.requestBody.Path)
					}
				}
			}

			// 에러 케이스인 경우 에러 메시지 확인
			if tt.expectError && tt.errorContains != "" {
				var response APIResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal error response: %v", err)
					return
				}

				if !strings.Contains(response.Error, tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, response.Error)
				}
			}
		})
	}
}

// TestValidateFileListRequest는 파일 목록 조회 요청 유효성 검사를 테스트
func TestValidateFileListRequest(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name          string
		request       FileListRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "유효한 요청",
			request: FileListRequest{
				Path:       "testdir",
				SortBy:     "name",
				SortOrder:  "asc",
				ShowHidden: false,
				MaxResults: 100,
			},
			expectError: false,
		},
		{
			name: "절대 경로 사용",
			request: FileListRequest{
				Path: "/etc/passwd",
			},
			expectError:   true,
			errorContains: "access to system directory",
		},
		{
			name: "잘못된 정렬 기준",
			request: FileListRequest{
				Path:   "testdir",
				SortBy: "invalid",
			},
			expectError:   true,
			errorContains: "invalid sort_by value",
		},
		{
			name: "잘못된 정렬 순서",
			request: FileListRequest{
				Path:      "testdir",
				SortOrder: "invalid",
			},
			expectError:   true,
			errorContains: "invalid sort_order value",
		},
		{
			name: "최대 결과 수 초과",
			request: FileListRequest{
				Path:       "testdir",
				MaxResults: 2000,
			},
			expectError:   true,
			errorContains: "max_results cannot exceed 1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateFileListRequest(&tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidateFileSearchRequest는 파일 검색 요청 유효성 검사를 테스트
func TestValidateFileSearchRequest(t *testing.T) {
	handler := NewAPIHandler()

	tests := []struct {
		name          string
		request       FileSearchRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "유효한 와일드카드 패턴",
			request: FileSearchRequest{
				Pattern: "*.txt",
			},
			expectError: false,
		},
		{
			name: "검색 조건 없음",
			request: FileSearchRequest{
				// 모든 검색 조건이 비어있음
			},
			expectError:   true,
			errorContains: "at least one search condition must be specified",
		},
		{
			name: "너무 광범위한 패턴",
			request: FileSearchRequest{
				Pattern: "*",
			},
			expectError:   true,
			errorContains: "too broad search pattern",
		},
		{
			name: "너무 긴 정규표현식",
			request: FileSearchRequest{
				RegexPattern: strings.Repeat("a", 101),
			},
			expectError:   true,
			errorContains: "regex pattern is too long",
		},
		{
			name: "너무 긴 확장자",
			request: FileSearchRequest{
				Extension: strings.Repeat("a", 21),
			},
			expectError:   true,
			errorContains: "extension is too long",
		},
		{
			name: "너무 긴 검색 내용",
			request: FileSearchRequest{
				Content: strings.Repeat("a", 201),
			},
			expectError:   true,
			errorContains: "search content is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateFileSearchRequest(&tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
