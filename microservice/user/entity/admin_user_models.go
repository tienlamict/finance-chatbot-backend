package entity

import (
	"finance-chatbot/addon/core"
	"strings"
)

// AdminCreateUserRequest represents a request from admin to create a new user
type AdminCreateUserRequest struct {
	FirstName  string     `json:"first_name" binding:"required"`
	LastName   string     `json:"last_name" binding:"required"`
	Email      string     `json:"email" binding:"required,email"`
	Phone      string     `json:"phone"`
	Gender     Gender     `json:"gender"`
	SystemRole SystemRole `json:"system_role"`
	Status     Status     `json:"status"`
}

func (req *AdminCreateUserRequest) Validate() error {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	if err := checkFirstName(req.FirstName); err != nil {
		return err
	}

	if err := checkLastName(req.LastName); err != nil {
		return err
	}

	if !emailIsValid(req.Email) {
		return ErrEmailIsNotValid
	}

	if req.Phone != "" {
		if err := checkPhoneNumber(req.Phone); err != nil {
			return err
		}
	}

	if req.Gender == "" {
		req.Gender = GenderUnknown
	} else {
		if err := checkGender(req.Gender); err != nil {
			return err
		}
	}

	if req.SystemRole == "" {
		req.SystemRole = RoleUser
	} else {
		if err := checkRole(req.SystemRole); err != nil {
			return err
		}
	}

	if req.Status == "" {
		req.Status = StatusActive
	} else {
		if err := checkStatus(req.Status); err != nil {
			return err
		}
	}

	return nil
}

// AdminUpdateUserRequest represents a request from admin to update a user
type AdminUpdateUserRequest struct {
	FirstName  *string     `json:"first_name"`
	LastName   *string     `json:"last_name"`
	Phone      *string     `json:"phone"`
	Gender     *Gender     `json:"gender"`
	SystemRole *SystemRole `json:"system_role"`
	Status     *Status     `json:"status"`
	Email      *string     `json:"email"`
}

func (req *AdminUpdateUserRequest) Validate() error {
	if req.FirstName != nil {
		s := strings.TrimSpace(*req.FirstName)
		if err := checkFirstName(s); err != nil {
			return err
		}
		req.FirstName = &s
	}

	if req.LastName != nil {
		s := strings.TrimSpace(*req.LastName)
		if err := checkLastName(s); err != nil {
			return err
		}
		req.LastName = &s
	}

	if req.Email != nil {
		s := strings.TrimSpace(*req.Email)
		if !emailIsValid(s) {
			return ErrEmailIsNotValid
		}
		req.Email = &s
	}

	if req.Phone != nil {
		s := strings.TrimSpace(*req.Phone)
		if s != "" {
			if err := checkPhoneNumber(s); err != nil {
				return err
			}
		}
		req.Phone = &s
	}

	if req.Gender != nil {
		if err := checkGender(*req.Gender); err != nil {
			return err
		}
	}

	if req.SystemRole != nil {
		if err := checkRole(*req.SystemRole); err != nil {
			return err
		}
	}

	if req.Status != nil {
		if err := checkStatus(*req.Status); err != nil {
			return err
		}
	}

	return nil
}

// ListUsersFilter represents filters for listing users
type ListUsersFilter struct {
	Search      string     `json:"search" form:"search"` // Search in name or email
	Role        SystemRole `json:"role" form:"role"`     // Filter by role
	Status      Status     `json:"status" form:"status"` // Filter by status
	Email       string     `json:"email" form:"email"`   // Filter by exact email
	core.Paging `json:",inline" form:",inline"`
}

func (f *ListUsersFilter) Process() {
	f.Paging.Process()
	f.Search = strings.TrimSpace(f.Search)
	f.Email = strings.TrimSpace(f.Email)
}

// UserListItem represents a user in the list response
type UserListItem struct {
	User
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Data   []UserListItem `json:"data"`
	Paging core.Paging    `json:"paging"`
}

// AdminCreateUserResponse represents the response when admin creates a user
type AdminCreateUserResponse struct {
	User              User   `json:"user"`
	TemporaryPassword string `json:"temporary_password"`
	Message           string `json:"message"`
}
