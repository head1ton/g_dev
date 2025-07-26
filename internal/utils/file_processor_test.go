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

// 파일 쓰기 기능 테스트
func TestFileProcessor_WriteFile(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	t.Run("write new file", func(t *testing.T) {
		// 새 파일 쓰기
		testContent := "Hello, World!\nThis is a test file."
		err := fp.WriteFile("test.txt", []byte(testContent))

		if err != nil {
			t.Errorf("Unexpeccted error: %v", err)
		}

		// 파일이 실제로 생성되었는지 확인
		writtenData, err := ioutil.ReadFile(filepath.Join(tempDir, "test.txt"))
		if err != nil {
			t.Errorf("Failed to read written file: %v", err)
		}

		if string(writtenData) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(writtenData))
		}

		// 히스토리 확인
		if len(fp.History) != 1 {
			t.Errorf("Expected 1 history entry, got %d", len(fp.History))
		}

		lastOp := fp.History[0]
		if lastOp.Operation != "write" {
			t.Errorf("Expected operation 'write', got '%s'", lastOp.Operation)
		}
		if !lastOp.Success {
			t.Error("Expected successful operation")
		}
		if lastOp.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), lastOp.Size)
		}
	})

	t.Run("write to subdirectory", func(t *testing.T) {
		// 하위 디렉토리에 파일 쓰기
		testContent := "Subdirectory test content"
		err := fp.WriteFile("subdir/test.txt", []byte(testContent))

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 파일이 실제로 생성되었는지 확인
		writtenData, err := ioutil.ReadFile(filepath.Join(tempDir, "subdir", "test.txt"))
		if err != nil {
			t.Errorf("Failed to read written file: %v", err)
		}

		if string(writtenData) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(writtenData))
		}
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		// 기존 파일 덮어쓰기
		originalContent := "Original content"
		newContent := "New content"

		// 첫 번째 파일 생성
		err := fp.WriteFile("overwrite.txt", []byte(originalContent))
		if err != nil {
			t.Errorf("Failed to write original file: %v", err)
		}

		// 파일 덮어쓰기
		err = fp.WriteFile("overwrite.txt", []byte(newContent))
		if err != nil {
			t.Errorf("Failed to overwrite file: %v", err)
		}

		// 덮어쓴 내용 확인
		writtenData, err := ioutil.ReadFile(filepath.Join(tempDir, "overwrite.txt"))
		if err != nil {
			t.Errorf("Failed to read overwritten file : %v", err)
		}

		if string(writtenData) != newContent {
			t.Errorf("Expected content '%s', got '%s'", newContent, string(writtenData))
		}
	})

	t.Run("write with absolute path", func(t *testing.T) {
		// 절대 경로로 파일 쓰기
		testContent := "Absolute path test"
		absPath := filepath.Join(tempDir, "absolute.txt")

		err := fp.WriteFile(absPath, []byte(testContent))
		if err != nil {
			t.Errorf("Failed to read absolute path file: %v", err)
		}

		// 파일 확인
		writtenData, err := ioutil.ReadFile(absPath)
		if err != nil {
			t.Errorf("Failed to read absolute path file: %v", err)
		}

		if string(writtenData) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(writtenData))
		}
	})
}

// 디렉토리 목록 기능을 테스트
func TestFileProcessor_ListDirectory(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// 테스트 파일들 생성
	testFiles := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		// 하위 디렉토리 생성
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		// 파일 생성
		if err := ioutil.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("list directory contents", func(t *testing.T) {
		// 디렉토리 목록 조회
		entries, err := fp.ListDirectory(".")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 최소 3개 항목이 있어야 함 (file1.txt, file2.txt, subdir)
		if len(entries) < 3 {
			t.Errorf("Expected at least 3 entires, got %d", len(entries))
		}

		// 파일과 디렉토리 확인
		foundFiles := 0
		foundDirs := 0
		for _, entry := range entries {
			if entry.IsDir() {
				foundDirs++
			} else {
				foundFiles++
			}
		}

		if foundFiles < 2 {
			t.Errorf("Expected at least 2 files, got %d", foundFiles)
		}
		if foundDirs < 1 {
			t.Errorf("Expected at least 1 directory, got %d", foundDirs)
		}

		// 히스토리 확인
		if len(fp.History) != 1 {
			t.Errorf("Expected 1 history entry, got %d", len(fp.History))
		}

		lastOp := fp.History[0]
		if lastOp.Operation != "list" {
			t.Errorf("Expected operation 'list', got '%s'", lastOp.Operation)
		}
		if !lastOp.Success {
			t.Error("Expected successful operation")
		}
	})

	t.Run("list non-existent directory", func(t *testing.T) {
		// 존재하지 않는 디렉토리 조회
		_, err := fp.ListDirectory("nonexistent")

		if err == nil {
			t.Error("Expected error for non-existent directory")
		}

		// 히스토리 확인
		if len(fp.History) != 2 {
			t.Errorf("Expected 2 history entries, got %d", len(fp.History))
		}

		lastOp := fp.History[1]
		if lastOp.Operation != "list" {
			t.Errorf("Expected operation 'list', got '%s'", lastOp.Operation)
		}
		if lastOp.Success {
			t.Error("Expected failed operation")
		}
	})
}

// 재귀 디렉토리 탐색 기능 테스트
func TestFileProcessor_WalkDirectory(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir := t.TempDir()
	fp := NewFileProcessor(tempDir)

	// 복잡한 디렉토리 구조 생성
	testStructure := []string{
		"file1.txt",
		"dir1/file2.txt",
		"dir1/subdir/file3.txt",
		"dir2/file4.txt",
	}

	for _, path := range testStructure {
		fullPath := filepath.Join(tempDir, path)
		// 디렉토리 생성
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		// 파일 생성
		if err := ioutil.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("walk directory recursively", func(t *testing.T) {
		// 재귀적으로 디렉토리 탐색
		files, err := fp.WalkDirectory(".")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 모든 파일이 포함되어야 함
		if len(files) != len(testStructure) {
			t.Errorf("Expected %d files, got %d", len(testStructure), len(files))
		}

		// 각 파일이 목록에 있는지 확인
		for _, expectedFile := range testStructure {
			found := false
			for _, file := range files {
				if file == expectedFile {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected file '%s' not found in walk results", expectedFile)
			}
		}

		// 히스토리 확인
		if len(fp.History) != 1 {
			t.Errorf("Expected 1 history entry, got %d", len(fp.History))
		}

		lastOp := fp.History[0]
		if lastOp.Operation != "walk" {
			t.Errorf("Expected operation 'walk', got '%s'", lastOp.Operation)
		}

		if !lastOp.Success {
			t.Error("Expected successful operation")
		}
	})

	t.Run("walk empty directory", func(t *testing.T) {
		// 빈 디렉토리 생성
		emptyDir := filepath.Join(tempDir, "empty")
		if err := os.MkdirAll(emptyDir, 0755); err != nil {
			t.Fatalf("Failed to create empty directory: %v", err)
		}

		// 빈 디렉토리 탐색
		files, err := fp.WalkDirectory("empty")

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 빈 디렉토리는 빈 슬라이스 반환
		if len(files) != 0 {
			t.Errorf("Expected empty slice, got %d files", len(files))
		}
	})
}
