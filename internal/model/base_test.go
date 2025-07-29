package model

import (
	"testing"
	"time"
)

func TestBaseModel_BeforeCreate(t *testing.T) {
	base := &BaseModel{}

	// BeforeCreate 호출
	err := base.BeforeCreate(nil)
	//log.Print(err)
	if err != nil {
		t.Errorf("BeforeCreate should not return error: %v", err)
	}

	// 생성 시간과 수정 시간이 설정되었는지 확인
	if base.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set after BeforeCreate")
	}

	if base.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after BeforeCreate")
	}

	// 생성 시간과 수정 시간이 동일한지 확인
	if !base.CreatedAt.Equal(base.UpdatedAt) {
		t.Error("CreatedAt and UpdatedAt should be equal after BeforeCreate")
	}
}

// BaseModel의 BeforeUpdate 훅 테스트
func TestBaseModel_BeforeUpdate(t *testing.T) {
	base := &BaseModel{}

	// 초기 시간 설정
	initialTime := time.Now().Add(-time.Hour)
	//log.Print(initialTime)
	base.CreatedAt = initialTime
	base.UpdatedAt = initialTime

	// BeforeUpdate 호출
	err := base.BeforeUpdate(nil)
	if err != nil {
		t.Errorf("BeforeUpdate should not return error: %v", err)
	}

	// 생성 시간은 변경되지 않았는지 확인
	if !base.CreatedAt.Equal(initialTime) {
		t.Error("CreatedAt should not be changed after BeforeUpdate")
	}

	// 수정 시간이 업데이트되었는지 확인
	if base.UpdatedAt.Equal(initialTime) {
		t.Error("UpdatedAt should be updated after BeforeUpdate")
	}

	// 수정 시간이 현재 시간과 가까운지 확인
	if time.Since(base.UpdatedAt) > time.Second {
		t.Error("UpdatedAt should be close to current time")
	}
}

// BaseModel의 IsDeleted 메서드를 테스트
func TestBaseModel_IsDeleted(t *testing.T) {
	base := &BaseModel{}

	// 삭제되지 않은 상태
	if base.IsDeleted() {
		t.Error("IsDeleted should return false when not deleted")
	}

	// 삭제된 상태 (소프트 상태)
	base.DeletedAt.Time = time.Now()
	if !base.IsDeleted() {
		t.Error("IsDeleted should return true when deleted")
	}
}

// BaseModel의 GetAge 메서드 테스트
func TestBaseModel_GetAge(t *testing.T) {
	base := &BaseModel{}

	// 생성 시간 설정 (1시간 전)
	base.CreatedAt = time.Now().Add(-time.Hour)

	age := base.GetAge()

	// 나이가 약 1시간인지 확인 (1초 오차 허용)
	expectedAge := time.Hour
	if age < expectedAge-time.Second || age > expectedAge+time.Second {
		t.Errorf("Expected age around %v, got %v", expectedAge, age)
	}
}

// GetLastModified 메서드 테스트
func TestBaseModel_GetLastModified(t *testing.T) {
	base := &BaseModel{}

	// 수정 시간 설정 (30분 전)
	base.UpdatedAt = time.Now().Add(-30 * time.Minute)

	lastModified := base.GetLastModified()

	// 마지막 수정이 약 30분 전인지 확인 (1초 오차 허용)
	expectedLastModified := 30 * time.Minute
	if lastModified < expectedLastModified-time.Second || lastModified > expectedLastModified+time.Second {
		t.Errorf("Expected last modified around %v, got %v", expectedLastModified, lastModified)
	}
}

// BeforeSave  Hook 테스트
func TestBaseModel_BeforeSave(t *testing.T) {
	base := &BaseModel{}

	// Validator를 구현하지 않은 경우
	err := base.BeforeSave(nil)
	if err != nil {
		t.Errorf("BeforeSave should not return error for non-validator: %v", err)
	}
}

// BaseModel의 사용 예제
func ExampleBaseModel() {
	base := &BaseModel{}

	// 생성 훅 호출
	base.BeforeCreate(nil)

	// 나이 확인
	age := base.GetAge()
	_ = age // 나이 사용

	// 마지막 수정 확인
	lastModified := base.GetLastModified()
	_ = lastModified // 마지막 수정 시간 사용

	// 삭제 상태 확인
	isDeleted := base.IsDeleted()
	_ = isDeleted // 삭제 상태 사용

}
