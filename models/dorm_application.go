package models

import (
	"time"

	"github.com/google/uuid"
)

type DormApplication struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	StudentID      string    `json:"student_id" gorm:"index"`            // mã sinh viên
	FullName       string    `json:"full_name"`                            // họ tên
	DOB            *time.Time `json:"dob,omitempty"`                       // ngày sinh
	Gender         string    `json:"gender,omitempty"`                    // giới tính
	CCCD           string    `json:"cccd,omitempty"`                      // CMND/CCCD
	CCCDIssueDate  *time.Time `json:"cccd_issue_date,omitempty"`          // ngày cấp
	CCCDIssuePlace string    `json:"cccd_issue_place,omitempty"`          // nơi cấp
	Phone          string    `json:"phone,omitempty"`                     // số điện thoại
	Email          string    `json:"email,omitempty"`                     // email
	AvatarFront    string    `json:"avatar_front,omitempty"`              // ảnh mặt trước (file path)
	AvatarBack     string    `json:"avatar_back,omitempty"`               // ảnh mặt sau (file path)
	Class          string    `json:"class,omitempty"`                     // lớp
	Course         string    `json:"course,omitempty"`                    // khóa
	Faculty        string    `json:"faculty,omitempty"`                   // ngành/ngành học
	Ethnicity      string    `json:"ethnicity,omitempty"`                 // dân tộc
	Religion       string    `json:"religion,omitempty"`                  // tôn giáo
	Hometown       string    `json:"hometown,omitempty"`                  // quê quán
	GuardianName   string    `json:"guardian_name,omitempty"`             // họ tên người bảo lãnh
	GuardianPhone  string    `json:"guardian_phone,omitempty"`            // SDT người bảo lãnh
	PriorityProof   string    `json:"priority_proof,omitempty"`            // minh chứng đối tượng ưu tiên (file path)
	PreferredSite  string    `json:"preferred_site,omitempty"`            // cơ sở (ví dụ: Hà Đông)
	PreferredDorm  string    `json:"preferred_dorm,omitempty"`            // ký túc xá mong muốn (ví dụ: B2)
	PriorityGroup  string    `json:"priority_group,omitempty"`            // đối tượng ưu tiên
	AdmissionType  string    `json:"admission_type,omitempty"`            // hệ đào tạo (chính quy...)
	Status         string    `json:"status" gorm:"default:'pending'"`   // trạng thái đơn: pending/approved/rejected
	Notes          string    `json:"notes,omitempty"`                     // ghi chú
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}