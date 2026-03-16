package dto

type Empty struct {}

type ErrorInternal struct {
	Code string `json:"code" example:"internal"`
	Error string `json:"error" example:"an error occured"`
}

type ResponseRegister struct {
	Status bool `json:"Status" example:"true"`
	Message string `json:"Message" example:"registered successfully"`
}

type ResponseLogin struct {
	Status bool `json:"Status" example:"true"`
	Message string `json:"Message" example:"successfully login"`
}

type ResponseRefresh struct {
	Status bool `json:"Status" example:"true"`
	Message string `json:"Message" example:"successfully refresh token"`
}

type ResponseLogout struct {
	Status bool `json:"Status" example:"true"`
	Message string `json:"Message" example:"successfully logout"`
}