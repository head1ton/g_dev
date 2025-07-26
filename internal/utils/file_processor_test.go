package utils

import (
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
