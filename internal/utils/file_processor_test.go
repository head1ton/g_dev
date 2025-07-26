package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// FileProcessor 생성자 테스트
func TestNewFileProcessor(t *testing.T) {
	testCases := []struct {
		name        string
		workingDir  string
		expectedDir string
	}{
		{"with working directory", "/test/path", "/test/path"},
		{"empty working directory", "", "."},
		{"current directory", ".", "."},
		{"parent directory", "..", ".."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fp := NewFileProcessor(tc.workingDir)

			if fp.WorkingDir != tc.expectedDir {
				t.Errorf("Expected working directory '%s', got '%s'", tc.expectedDir, fp.WorkingDir)
			}

			if fp.History == nil {
				t.Error("History slice should be initialized, not nil")
			}

			if len(fp.History) != 0 {
				t.Errorf("Expected empty history, got length %d", len(fp.History))
			}
		})
	}
}

// 작업 디렉토리 기능을 테스트
func TestFileProcessor_WorkingDirectory(t *testing.T) {
	// 현재 디렉토리 가져오기
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// 현재 디렉토리로 FileProcessor 생성
	fp := NewFileProcessor(currentDir)

	// 작업 디렉토리가 정확한지 확인
	if fp.WorkingDir != currentDir {
		t.Errorf("Expected working directory '%s', got '%s'", currentDir, fp.WorkingDir)
	}

	// 상대 경로를 절대 경로로 변환하는 기능 테스트
	absPath := filepath.Join(currentDir, "test.txt")
	if filepath.IsAbs(absPath) != true {
		t.Errorf("Expected absolute path, got relative path: %s", absPath)
	}
}

// 파일 읽기 기능 테스트
func TestFileProcessor_ReadFile(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)
	//t.Logf("Temp dir: %s", tempDir)
	//t.Log(fp)

	// 테스트 파일 생성
	testContent := "Hello, World!\nThis is a test file."
	testFile := filepath.Join(tempDir, "test.txt")

	err := ioutil.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("read existing file", func(t *testing.T) {
		// 파일 읽기
		data, err := fp.ReadFile("test.txt")

		if err != nil {
			t.Errorf("Unexpected erro: %v", err)
		}

		if string(data) != testContent {
			t.Errorf("Exprected content '%s', got '%s'", testContent, string(data))
		}

		// 히스토리 확인
		if len(fp.History) != 1 {
			t.Errorf("Expected 1 history entry, got %d", len(fp.History))
		}

		lastOp := fp.History[0]
		if lastOp.Operation != "read" {
			t.Errorf("Expected operation 'read', got '%s'", lastOp.Operation)
		}

		if !lastOp.Success {
			t.Error("Expected successful operation")
		}
		if lastOp.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), lastOp.Size)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		// 존재하지 않는 파일 읽기 시도
		_, err := fp.ReadFile("nonexistent.txt")

		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		// 히스토리 확인
		if len(fp.History) != 2 {
			t.Errorf("Expected 2 history entries, got %d", len(fp.History))
		}

		lastOp := fp.History[1]
		if lastOp.Operation != "read" {
			t.Errorf("Expected operation 'read', got '%s'", lastOp.Operation)
		}
		if lastOp.Success {
			t.Error("Expected failed operation")
		}
		if lastOp.Error == "" {
			t.Error("Expected error message")
		}
	})

	t.Run("read with absolute path", func(t *testing.T) {
		// 절대 경로로 파일 읽기
		data, err := fp.ReadFile(testFile)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if string(data) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(data))
		}
	})
}
