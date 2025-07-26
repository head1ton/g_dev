package utils

import "testing"

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
