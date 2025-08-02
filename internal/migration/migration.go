package migration

import (
	"fmt"
	"g_dev/internal/database"
	"g_dev/internal/model"
	"log"
)

// 데이터베이스 마이그레이션을 관리
type MigrationManager struct {
	// 데이터베이스 인스턴스
	DB *database.Database

	// 등록된 모델들
	Models []interface{}

	// 마이그레이션 상태
	IsMigrated bool
}

// 새로운 MigrationManager 인스턴스 생성
func NewMigrationManager(db *database.Database) *MigrationManager {
	return &MigrationManager{
		DB:         db,
		Models:     make([]interface{}, 0),
		IsMigrated: false,
	}
}

// 마이그레이션 할 모델 등록
func (m *MigrationManager) RegisterModel(model interface{}) {
	m.Models = append(m.Models, model)
}

// 기본 모델들을 등록
func (m *MigrationManager) RegisterDefaultModels() {
	// 사용자 관련 모델
	m.RegisterModel(&model.User{})

	// 게임 관련 모델
	m.RegisterModel(&model.Game{})

	// 점수 관련 모델
	m.RegisterModel(&model.Score{})

	// 인벤토리 관련 모델
	m.RegisterModel(&model.Inventory{})

	// 추가 모델 등록
}

// 등록된 모든 모델 마이그레이션
func (m *MigrationManager) Migrate() error {
	if len(m.Models) == 0 {
		return fmt.Errorf("마이그레이션할 모델이 등록되지 않았습니다")
	}

	log.Printf("데이터베이스 마이그레이션 시작: %d개 모델", len(m.Models))

	// 각 모델을 마이그레이션
	for i, model := range m.Models {
		log.Printf("    - 모델 %d/%d 마이그레이션 중...", i+1, len(m.Models))

		if err := m.DB.GetDB().AutoMigrate(model); err != nil {
			return fmt.Errorf("모델 마이그레이션 실패: %v", err)
		}
	}

	m.IsMigrated = true
	log.Printf("데이터베이스 마이그레이션 완료: %d개 모델", len(m.Models))

	return nil
}

// 등록된 모델 목록 반환
func (m *MigrationManager) GetRegisteredModels() []interface{} {
	return m.Models
}

// 마이그레이션이 완료되었는지 확인
func (m *MigrationManager) IsMigrationComplete() bool {
	return m.IsMigrated
}

// 마이그레이션 상태 초기화
func (m *MigrationManager) ResetMigrationStatus() {
	m.IsMigrated = false
}
