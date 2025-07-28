package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// 파일 목록 조회 API 요청
type FileListRequest struct {
	Path       string `json:"path" example:"."`            // 조회할 디렉토리 경로 (기본값: 현재 디렉토리)
	SortBy     string `json:"sort_by" example:"name"`      // 정렬 기준 (name, size, modified)
	SortOrder  string `json:"sort_order" example:"asc"`    // 정렬 순서 (asc, desc)
	ShowHidden bool   `json:"show_hidden" example:"false"` // 숨김 파일 표시 여부
	MaxResults int    `json:"max_results" example:"100"`   // 최대 결과 수 (기본값 100)
}

// 파일/디렉토리 정보를 담는 구조체
type FileInfo struct {
	Name         string `json:"name"`          // 파일/디렉토리 이름
	Path         string `json:"path"`          // 전체 경로
	Size         int64  `json:"size"`          // 파일 크기 (바이트)
	IsDirectory  bool   `json:"is_directory"`  // 디렉토리 여부
	IsHidden     bool   `json:"is_hidden"`     // 숨김 파일 여부
	Permissions  string `json:"permissions"`   // 파일 권한 (예: -rw-r--r--)
	ModifiedTime string `json:"modified_time"` // 수정 시간 (ISO 8601 형식)
	Extension    string `json:"extension"`     // 파일 확장자(파일인 경우)
}

// 파일 목록 조회 API 응답
type FileListResponse struct {
	Path        string                 `json:"path"`        // 조회된 디렉토리 경로
	Files       []FileInfo             `json:"files"`       // 파일 목록
	Directories []FileInfo             `json:"directories"` // 디렉토리 목록
	TotalCount  int                    `json:"total_count"` // 전체 항목 수
	Summary     map[string]interface{} `json:"summary"`     // 요약 정보 (총 크기, 파일 수 등)
}

// 파일 읽기 API 응답
type FileReadResponse struct {
	Path      string `json:"path"`       // 읽은 파일 경로
	Content   string `json:"content"`    // 파일 내용
	Size      int64  `json:"size"`       // 파일 크기
	LineCount int    `json:"line_count"` // 줄 수
	Encoding  string `json:"encoding"`   // 사용된 인코딩
	ReadTime  string `json:"read_time"`  // 읽기 완료 시간
}

// 파일 쓰기 API 요청
type FileWriteRequest struct {
	Path      string `json:"path" example:"output.txt"`     // 쓸 파일 경로
	Content   string `json:"content" example:"Hello World"` // 파일 내용
	Encoding  string `json:"encoding" example:"utf-8"`      // 파일 인코딩
	Append    bool   `json:"append" example:"false"`        // 추가 모드 여부 (기본값: 덮어쓰기)
	CreateDir bool   `json:"create_dir" example:"true"`     // 디렉토리 자동 생성 여부
}

// 파일 쓰기 API 응답
type FileWriteResponse struct {
	Path      string `json:"path"`       // 쓴 파일 경로
	Size      int64  `json:"size"`       // 파일 크기
	LineCount int    `json:"line_count"` // 줄 수
	WriteTime string `json:"write_time"` // 쓰기 완료 시간
	Created   bool   `json:"created"`    // 새로 생성된 파일 여부
}

// 파일 목록 조회 API
// 지정된 디렉토리의 파일과 폴더 목록 반환
// @Summary 파일 목록 조회
// @Description 지정된 디렉토리의 파일과 폴더 목록을 조회합니다.
// @Tags FileProcessor
// @Accept json
// @Produce json
// @Param request body FileListRequest true "목록 조회 요청"
// @Success 200 {object} APIResponse{data=FileListResponse} "목록 조회 성공"
// @Failure 400 {object} APIResponse "잘못된 요청"
// @Failure 403 {object} APIResponse "접근 권한 없음"
// @Failure 404 {object} APIResponse "디렉토리를 찾을 수 없음"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/files/list [post]
func (h *APIHandler) HandleFileList(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST")
		return
	}

	// 요청 본문 파싱
	var request FileListRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON request: "+err.Error())
		return
	}

	// 요청 유효성 검사
	if err := h.validateFileListRequest(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 파일 목록 조회
	response, err := h.performFileList(&request)
	if err != nil {
		// 에러 타입에 따른 적절한 HTTP 상태 코드 설정
		if strings.Contains(err.Error(), "permission denied") {
			h.writeErrorResponse(w, http.StatusForbidden, "Access denied: "+err.Error())
		} else if strings.Contains(err.Error(), "no such file") {
			h.writeErrorResponse(w, http.StatusNotFound, "Directory not found: "+err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "File list error: "+err.Error())
		}
		return
	}

	// 성공 응답
	h.writeSuccessResponse(w, response, "파일 목록 조회가 완료되었습니다")
}

// HandleFileSearch는 파일 검색 API 엔드포인트를 처리합니다.
// 다양한 조건으로 파일을 검색합니다.
// @Summary 파일 검색
// @Description 다양한 조건(패턴, 확장자, 내용 등)으로 파일을 검색합니다.
// @Tags FileProcessor
// @Accept json
// @Produce json
// @Param request body FileSearchRequest true "검색 요청"
// @Success 200 {object} APIResponse{data=[]string} "검색 성공"
// @Failure 400 {object} APIResponse "잘못된 요청"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/files/search [post]
func (h *APIHandler) HandleFileSearch(w http.ResponseWriter, r *http.Request) {
	// HTTP 메서드 검증 (POST만 허용)
	if r.Method != http.MethodPost {
		h.writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed. Use POST")
		return
	}

	// 요청 본문 파싱
	var request FileSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON request: "+err.Error())
		return
	}

	// 요청 유효성 검사
	if err := h.validateFileSearchRequest(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// 파일 검색 수행
	results, err := h.performFileSearch(&request)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "File search error: "+err.Error())
		return
	}

	// 성공 응답
	h.writeSuccessResponse(w, results, "파일 검색이 완료되었습니다")
}

// HandleFileRead는 파일 읽기 API 엔드포인트를 처리합니다.
// 지정된 파일의 내용을 읽어서 반환합니다.
// @Summary 파일 읽기
// @Description 지정된 파일의 내용을 읽어서 반환합니다.
// @Tags FileProcessor
// @Accept json
// @Produce json
// @Param request body FileReadRequest true "파일 읽기 요청"
// @Success 200 {object} APIResponse{data=FileReadResponse} "파일 읽기 성공"
// @Failure 400 {object} APIResponse "잘못된 요청"
// @Failure 403 {object} APIResponse "접근 권한 없음"
// @Failure 404 {object} APIResponse "파일을 찾을 수 없음"
// @Failure 405 {object} APIResponse "허용되지 않는 HTTP 메서드"
// @Router /api/files/read [post]
func (h *APIHandler) HandleFileRead(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// 파일 목록 조회 요청의 유효성 검사
// 경로 유효성, 정렬 옵션, 결과 수 제한 등을 확인
func (h *APIHandler) validateFileListRequest(request *FileListRequest) error {
	// 경로 유효성 검사
	if request.Path == "" {
		request.Path = "." // 기본값: 현재 디렉토리
	}

	// 절대 경로 보안 검사 (상위 디렉토리 접근 방지)
	// 테스트 환경에서는 임시 디렉토리 경로를 허용
	if filepath.IsAbs(request.Path) {
		// 시스템 중요 디레고리 접근 방지
		restrictedPaths := []string{"/etc", "/var", "/usr", "/bin", "/sbin", "/boot", "/dev", "/proc", "/sys"}
		for _, restricted := range restrictedPaths {
			if strings.HasPrefix(request.Path, restricted) {
				return fmt.Errorf("access to system directory '%s' is not allowed for security reasons", restricted)
			}
		}
	}

	// 정렬 기준 검사
	validSortBy := map[string]bool{
		"name":     true,
		"size":     true,
		"modified": true,
	}
	if request.SortBy != "" && !validSortBy[request.SortBy] {
		return fmt.Errorf("invalid sort_by value. Valid values: name, size, modified")
	}

	// 정렬 순서 검사
	validSortOrder := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if request.SortOrder != "" && !validSortOrder[request.SortOrder] {
		return fmt.Errorf("invalid sort_order value. Valid values: asc, desc")
	}

	// 최대 결과 수 제한
	if request.MaxResults <= 0 {
		request.MaxResults = 100 // 기본값
	} else if request.MaxResults > 1000 {
		return fmt.Errorf("max_results cannot exceed 1000 for performance reasons")
	}

	return nil
}

// 파일 검색 요청의 유효성 검사
// 검색 조건의 유효성과 보안을 확인합니다.
func (h *APIHandler) validateFileSearchRequest(request *FileSearchRequest) error {
	// 최소한 하나의 검색 조건이 있어야 함
	if request.Pattern == "" && request.RegexPattern == "" &&
		request.Extension == "" && request.Content == "" {
		return fmt.Errorf("at least one search condition must be specified")
	}

	// 와일드카드 패턴 검사
	if request.Pattern != "" {
		// 위험한 패턴 방지 (예: `*`만 있는 경우)
		if request.Pattern == "*" || request.Pattern == "*.*" {
			return fmt.Errorf("too broad search pattern is not allowed for performance reasons")
		}
	}

	// 정규표현식 패턴 검사
	if request.RegexPattern != "" {
		// 복잡한 정규표현식으로 인한 성능 저하 방지
		if len(request.RegexPattern) > 100 {
			return fmt.Errorf("regex pattern is too long (max 100 characters)")
		}
	}

	// 확장자 검사
	if request.Extension != "" {
		// 확장자는 점(.)으로 시작하지 않아야 함
		if strings.HasPrefix(request.Extension, ".") {
			request.Extension = strings.TrimPrefix(request.Extension, ".")
		}
		// 확장자 길이 제한
		if len(request.Extension) > 20 {
			return fmt.Errorf("extension is too long (max 20 characters)")
		}
	}

	// 내용 검색 텍스트 검사
	if request.Content != "" {
		// 검색 텍스트 길이 제한
		if len(request.Content) > 200 {
			return fmt.Errorf("search content is too long (max 200 characters)")
		}
	}

	return nil
}

// 실제 파일 목록 조회를 수행
// FileProcessor를 사용하여 디렉토리 내용을 조회하고 응답을 구성
func (h *APIHandler) performFileList(request *FileListRequest) (*FileListResponse, error) {
	response := &FileListResponse{
		Path:        request.Path,
		Files:       []FileInfo{},
		Directories: []FileInfo{},
		TotalCount:  0,
		Summary: map[string]interface{}{
			"total_files":       0,
			"total_directories": 0,
			"total_size":        0,
		},
	}
	return response, nil
}

// 실제 파일 검색을 수행
// FileProcessor의 검색 메서드들을 사용하여 파일을 찾음
func (h *APIHandler) performFileSearch(request *FileSearchRequest) ([]string, error) {
	var results []string
	var err error

	// 검색 조건에 따라 적절한 검색 메서드 호출
	if request.Pattern != "" {
		results, err = h.fileProcessor.SearchFiles(request.Pattern)
	} else if request.RegexPattern != "" {
		results, err = h.fileProcessor.SearchFilesByRegex(request.RegexPattern)
	} else if request.Extension != "" {
		results, err = h.fileProcessor.SearchFilesByExtension(request.Extension)
	} else if request.Content != "" {
		results, err = h.fileProcessor.SearchFilesByContent(request.Content, request.CaseSensitive)
	}

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return results, nil
}
