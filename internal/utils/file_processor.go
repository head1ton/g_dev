package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// 파일 작업 결과를 저장하는 구조체
type FileOperation struct {
	Operation string `json:"operation"` // 수행된 작업 (read, write, list, etc)
	Path      string `json:"path"`      // 파일 디렉토리 경로
	Success   bool   `json:"success"`   // 작업 성공 여부
	Error     string `json:"error"`     // 에러 메시지
	Size      int64  `json:"size"`      // 파일 크기 (바이트)
}

type FileProcessor struct {
	// 작업 디렉토리 (기본 경로)
	WorkingDir string
	// 파일 처리 히스토리
	History []FileOperation
}

// 새로운 FileProcessor 인스턴스 생성
func NewFileProcessor(workingDir string) *FileProcessor {
	// 작업 디렉토리가 비어있으면 현재 디렉토리 사용
	if workingDir == "" {
		workingDir = "."
	}

	return &FileProcessor{
		WorkingDir: workingDir,
		History:    make([]FileOperation, 0),
	}
}

// 지정된 파일을 읽어서 내용을 반환. 파일이 존재하지 않거나 읽을 수 없는 경우 에러 반환
func (fp *FileProcessor) ReadFile(filename string) ([]byte, error) {
	// 절대 경로가 아니면 작업 디렉토리와 결합
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(fp.WorkingDir, filename)
	}

	// 파일 열기
	file, err := os.Open(filename)
	if err != nil {
		// 에러 발생 시 히스토리에 기록
		op := FileOperation{
			Operation: "read",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return nil, err
	}
	defer file.Close()

	// 파일 정보 가져오기
	fileInfo, err := file.Stat()
	if err != nil {
		op := FileOperation{
			Operation: "read",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return nil, err
	}

	// 파일 내용 읽기
	data, err := ioutil.ReadAll(file)
	if err != nil {
		op := FileOperation{
			Operation: "read",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return nil, err
	}

	// 성공 시 히스토리 기록
	op := FileOperation{
		Operation: "read",
		Path:      filename,
		Success:   true,
		Error:     "",
		Size:      fileInfo.Size(),
	}
	fp.History = append(fp.History, op)

	return data, nil
}

// 지정된 파일에 데이터를 쓰기. 파일이 존재하지 않으면 새로 생성. 존재하면 덮어씀
func (fp *FileProcessor) WriteFile(filename string, data []byte) error {
	// 절대 경로가 아니면 작업 디렉토리와 결합
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(fp.WorkingDir, filename)
	}

	// 디렉토리가 존재하지 않으면 생성
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		op := FileOperation{
			Operation: "write",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return err
	}

	// 파일 생성 (기존 파일 덮어쓰기)
	file, err := os.Create(filename)
	if err != nil {
		op := FileOperation{
			Operation: "write",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return err
	}
	defer file.Close()

	// 데이터 쓰기
	bytesWritten, err := file.Write(data)
	if err != nil {
		op := FileOperation{
			Operation: "write",
			Path:      filename,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return err
	}

	// 성공 시 히스토리에 기록
	op := FileOperation{
		Operation: "write",
		Path:      filename,
		Success:   true,
		Error:     "",
		Size:      int64(bytesWritten),
	}
	fp.History = append(fp.History, op)

	return nil
}

// 지정된 디렉토리의 내용을 반환
func (fp *FileProcessor) ListDirectory(dirPath string) ([]os.FileInfo, error) {
	// 절대 경로가 아니면 작업 디렉토리와 결합
	if !filepath.IsAbs(dirPath) {
		dirPath = filepath.Join(fp.WorkingDir, dirPath)
	}

	// 디렉토리 읽기
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		op := FileOperation{
			Operation: "list",
			Path:      dirPath,
			Success:   false,
			Error:     err.Error(),
			Size:      0,
		}
		fp.History = append(fp.History, op)
		return nil, err
	}

	// FileInfo 슬라이스로 변환
	var fileInfos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, info)
	}

	// 성공 시 히스토리에 기록
	op := FileOperation {
		Operation: "list",
		Path:      dirPath,
		Success:   true,
		Error: "",
		Size: int64(len(fileInfos)),
	}
	fp.History = append(fp.History, op)

	return fileInfos, nil
}

// 디렉토리를 재귀적으로 탐색
func (fp *FileProcessor) WalkDirectory(rootPath string) ([]string, error) {
	// 절대 경로가 아니면 디렉토리와 결합
	if !filepath.IsAbs(rootPath) {
		rootPath = filepath.Join(fp.WorkingDir, rootPath)
	}

	var allFiles []string

	// 재귀적으로 디렉토리 탐색
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err	// 탐색 중 오류 발생 시 중단
		}

		// 루트 디렉토리는 제외
		if path == rootPath {
			return nil
		}

		// 파일만 포함 (디렉토리 제외)
		if info.IsDir() {
			return nil
		}

		// 상대 경로만 변환
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		allFiles = append(allFiles, relPath)
		return nil
	})

	if err != nil {
		op := FileOperation {
			Operation: "walk",
			Path: rootPath,,
			Success: false,
			Error: err.Error(),
			Size: 0,,
		}
		fp.History = append(fp.History, op)
		return nil, err
	}

	// 성공 시 히스토리에 기록
	op := FileOperation{
		Operation: "walk",
		Path:      rootPath,
		Success:   true,
		Error:     "",
		Size:      int64(len(allFiles)),
	}
	fp.History = append(fp.History, op)

	return allFiles, nil
}
