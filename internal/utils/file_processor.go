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
