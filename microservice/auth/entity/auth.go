package entity

type Auth struct {
	UserName string
	AuthType string
	PassWord string
	Email    string
}

func (Auth) TableName() string { return "auths" }
