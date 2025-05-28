package auth

type LoginRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	AccessToken          string `json:"accessToken"`
	PassrefreshTokenword string `json:"passrefreshTokenword"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type ActivationRequest struct {
	// From ActivationInitiateRequest
	PhoneNumber string `json:"phoneNumber"`
	AccountNo   string `json:"accountNo"`
	NIK         string `json:"nik"`
	BirthDate   string `json:"birthDate"`
	MotherName  string `json:"motherName"`

	// From ActivationCompleteRequest
	Password     string `json:"password"`
	FullName     string `json:"fullName"`
	NickName     string `json:"nickName"`
	BirthPlace   string `json:"birthPlace"`
	Gender       string `json:"gender"`
	Religion     string `json:"religion"`
	Address      string `json:"address"`
	RT           string `json:"rt"`
	RW           string `json:"rw"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	SubDistrict  string `json:"subDistrict"`
	PostalCode   string `json:"postalCode"`
	NPWP         string `json:"npwp"`
	Email        string `json:"email"`
	Occupation   string `json:"occupation"`
	FundPurpose  string `json:"fundPurpose"`
	FundSource   string `json:"fundSource"`
	AnnualIncome string `json:"annualIncome"`
}

type OtpRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	OtpCode     string `json:"otpCode"`
}

type ProfileResponse struct {
	PhoneNumber string `json:"phoneNumber"`
	UserId      string `json:"userId"`
}
