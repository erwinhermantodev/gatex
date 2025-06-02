package auth

// LoginRequest represents the login credentials
type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,min=10,max=15"`
	Password    string `json:"password"`
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"` // Fixed typo from "PassrefreshTokenword"
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// OtpRequest represents OTP verification request
type OtpRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,min=10,max=15"`
	OtpCode     string `json:"otpCode"`
}

type CheckPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,min=10,max=15"`
}

// ProfileResponse represents user profile information
type ProfileResponse struct {
	PhoneNumber string `json:"phoneNumber"`
	UserId      string `json:"userId"`
}

// ActivationRequest represents account activation request
type ActivationRequest struct {
	// Basic Info
	PhoneNumber string `json:"phoneNumber" `
	AccountNo   string `json:"accountNo" `
	NIK         string `json:"nik" `
	BirthDate   string `json:"birthDate" `
	MotherName  string `json:"motherName" `

	// Personal Details
	Password     string `json:"password" `
	ReferralCode string `json:"referralCode"`
	FullName     string `json:"fullName" `
	NickName     string `json:"nickName"`
	BirthPlace   string `json:"birthPlace" `
	Gender       string `json:"gender" `
	Religion     string `json:"religion"`

	// Address Info
	Address     string `json:"address" `
	RT          string `json:"rt"`
	RW          string `json:"rw"`
	Province    string `json:"province" `
	City        string `json:"city" `
	District    string `json:"district" `
	SubDistrict string `json:"subDistrict" `
	PostalCode  string `json:"postalCode" `

	// Additional Info
	NPWP         string `json:"npwp" `
	Email        string `json:"email" `
	Occupation   string `json:"occupation"`
	FundPurpose  string `json:"fundPurpose"`
	FundSource   string `json:"fundSource"`
	AnnualIncome string `json:"annualIncome"`
}
