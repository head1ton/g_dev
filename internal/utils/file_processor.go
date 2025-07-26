package utils

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
