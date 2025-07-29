package model

import (
	"gorm.io/gorm"
	"time"
)

// 모든 모델이 공통으로 사용하는 기본 구조체
// ID, 생성 시간, 수정 시간, 삭제 시간을 포함
type BaseModel struct {
	// 기본 키 (자동 증가)
	ID uint `json:"id" gorm:"primarykey:autoIncrement"`

	// 생성 시간(레코드가 처음 생성된 시간)
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;not null"`

	// 수정 시간 (레코드가 마지막으로 수정된 시간)
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;not null"`

	// 삭제 시간 (소프트 삭제를 위한 필드, null 이면 삭제되지 않음)
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// 레코드 생성 전에 호출되는 GORM Hook
// 생성 시간을 현재 시간으로 설정
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	return nil
}

// 레코드 수정 전에 호출되는 GORM HOOK
// 수정 시간을 현재 시간으로 설정
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now()
	return nil
}

// 레코드가 삭제되었는지 확인
// 소프트 삭제된 경우 true를 반환
func (b *BaseModel) IsDeleted() bool {
	return !b.DeletedAt.Time.IsZero()
}

// 레코드 생성 후 경과된 시간을 반환
// 현재 시간에서 생성 시간을 뺀 값 반환
func (b *BaseModel) GetAge() time.Duration {
	return time.Since(b.CreatedAt)
}

// 마지막 수정 후 경과된 시간을 반환
// 현재 시간에서 수정 시간을 뺀 값을 반환
func (b *BaseModel) GetLastModified() time.Duration {
	return time.Since(b.UpdatedAt)
}

// 테이블 이름을 반환하는 인터페이스
// 각 모델에서 구현하여 커스텀 테이블 이름을 지정.
type TableName interface {
	TableName() string
}

// 데이터 유효성 감사를 위한 인터페이스
// 각 모델에서 구현하여 데이터 검증 로직을 추가
type Validator interface {
	Validate() error
}

// 레코드 저장 전에 호출되는 GORM HOOK
// Validator 인터페이스를 구현한 모델의 경우 유효성 검사 수행
func (b *BaseModel) BeforeSave(tx *gorm.DB) error {
	if validator, ok := interface{}(b).(Validator); ok {
		return validator.Validate()
	}
	return nil
}
