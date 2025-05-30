package auth

// LoginRequest represents the login credentials
type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,min=10,max=15"`
	Password    string `json:"password" validate:"required,min=6"`
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
	OtpCode     string `json:"otpCode" validate:"required,len=6"`
}

// ProfileResponse represents user profile information
type ProfileResponse struct {
	PhoneNumber string `json:"phoneNumber"`
	UserId      string `json:"userId"`
}

// ActivationRequest represents account activation request
type ActivationRequest struct {
	// Basic Info
	PhoneNumber string `json:"phoneNumber" validate:"required,min=10,max=15"`
	AccountNo   string `json:"accountNo" validate:"required"`
	NIK         string `json:"nik" validate:"required,len=16"`
	BirthDate   string `json:"birthDate" validate:"required"`
	MotherName  string `json:"motherName" validate:"required"`

	// Personal Details
	Password     string `json:"password" validate:"required,min=6"`
	ReferralCode string `json:"referralCode"`
	FullName     string `json:"fullName" validate:"required"`
	NickName     string `json:"nickName"`
	BirthPlace   string `json:"birthPlace" validate:"required"`
	Gender       string `json:"gender" validate:"required,oneof=M F"`
	Religion     string `json:"religion"`

	// Address Info
	Address     string `json:"address" validate:"required"`
	RT          string `json:"rt"`
	RW          string `json:"rw"`
	Province    string `json:"province" validate:"required"`
	City        string `json:"city" validate:"required"`
	District    string `json:"district" validate:"required"`
	SubDistrict string `json:"subDistrict" validate:"required"`
	PostalCode  string `json:"postalCode" validate:"required,len=5"`

	// Additional Info
	NPWP         string `json:"npwp" validate:"len=15"`
	Email        string `json:"email" validate:"email"`
	Occupation   string `json:"occupation"`
	FundPurpose  string `json:"fundPurpose"`
	FundSource   string `json:"fundSource"`
	AnnualIncome string `json:"annualIncome"`
}
