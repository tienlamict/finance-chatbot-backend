package entity

// SetPasswordRequest allows user or admin to set/reset password
type SetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func (req *SetPasswordRequest) Validate() error {
	if err := checkPassword(req.NewPassword); err != nil {
		return err
	}
	return nil
}

// AdminSetUserPasswordRequest allows admin to set user password
type AdminSetUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
	SendEmail   bool   `json:"send_email"` // Whether to email user the new password
}

func (req *AdminSetUserPasswordRequest) Validate() error {
	if err := checkPassword(req.NewPassword); err != nil {
		return err
	}
	return nil
}
