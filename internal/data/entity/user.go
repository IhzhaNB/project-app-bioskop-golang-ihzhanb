package entity

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	Base
	Username      string   `db:"username"`
	Email         string   `db:"email"`
	PasswordHash  string   `db:"password"`
	Phone         *string  `db:"phone"`
	Role          UserRole `db:"role"`
	EmailVerified bool     `db:"email_verified"`
	IsActive      bool     `db:"is_active"`
}
