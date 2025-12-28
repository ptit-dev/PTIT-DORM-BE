package models

import (
	"mime/multipart"
)

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         LoginUserInfo `json:"user"`
}

type LoginUserInfo struct {
	UserID      string   `json:"user_id"`
	Avatar      string   `json:"avatar"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

type Account struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	Roles     []string `json:"roles"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type ProfileStudentResponse struct {
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Student  Student  `json:"student"`
	Parents  []Parent `json:"parents"`
}

type ProfileManagerResponse struct {
	Email    string  `json:"email"`
	Username string  `json:"username"`
	Manager  Manager `json:"manager"`
}

type LogoutAllSessionsRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Dùng cho API nhận multipart/form-data khi tạo application
// Các trường file là *multipart.FileHeader, các trường text là string
// Không cần gorm tag vì không lưu trực tiếp struct này vào DB
type DormApplicationCreateRequest struct {
	StudentID      string                `form:"student_id"`
	FullName       string                `form:"full_name"`
	DOB            string                `form:"dob"` // yyyy-MM-dd
	Gender         string                `form:"gender"`
	CCCD           string                `form:"cccd"`
	CCCDIssueDate  string                `form:"cccd_issue_date"`
	CCCDIssuePlace string                `form:"cccd_issue_place"`
	Phone          string                `form:"phone"`
	Email          string                `form:"email"`
	AvatarFront    *multipart.FileHeader `form:"avatar_front"`
	AvatarBack     *multipart.FileHeader `form:"avatar_back"`
	Class          string                `form:"class"`
	Course         string                `form:"course"`
	Faculty        string                `form:"faculty"`
	Ethnicity      string                `form:"ethnicity"`
	Religion       string                `form:"religion"`
	Hometown       string                `form:"hometown"`
	GuardianName   string                `form:"guardian_name"`
	GuardianPhone  string                `form:"guardian_phone"`
	PriorityProof  *multipart.FileHeader `form:"priority_proof"`
	PreferredSite  string                `form:"preferred_site"`
	PreferredDorm  string                `form:"preferred_dorm"`
	PriorityGroup  string                `form:"priority_group"`
	AdmissionType  string                `form:"admission_type"`
	Notes          string                `form:"notes"`
}

type StaffWithUserInfo struct {
	StaffID    string `json:"staff_id"`
	FullName   string `json:"fullname"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	CCCD       string `json:"cccd"`
	Avatar     string `json:"avatar"`
	Province   string `json:"province"`
	Commune    string `json:"commune"`
	DetailAddr string `json:"detail_address"`
	DOB        string `json:"dob"`
}

// Thông tin nội trú lấy từ hợp đồng đã được duyệt
type ResidentInfo struct {
	Username  string `json:"username"`
	FullName  string `json:"fullname"`
	Class     string `json:"class"`
	Avatar    string `json:"avatar"`
	StudentID string `json:"student_id"`
}

// Thông tin hợp đồng đơn giản dùng để thống kê số lượng người / phòng
type ApprovedContractSummary struct {
	ContractID string `json:"contract_id"`
	Room       string `json:"room"`
}

// Thông tin yêu cầu đổi phòng kèm username 2 sinh viên
type RoomTransferRequestWithUsernames struct {
	ID                   string `json:"id"`
	RequesterUserID      string `json:"requester_user_id"`
	RequesterUsername    string `json:"requester_username"`
	TargetUserID         string `json:"target_user_id"`
	TargetUsername       string `json:"target_username"`
	TargetRoomID         string `json:"target_room_id"`
	TransferTime         string `json:"transfer_time"`
	Reason               string `json:"reason"`
	PeerConfirmStatus    string `json:"peer_confirm_status"`
	ManagerConfirmStatus string `json:"manager_confirm_status"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}
