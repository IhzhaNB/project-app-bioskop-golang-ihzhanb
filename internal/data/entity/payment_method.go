package entity

type PaymentMethod struct {
	Base
	Name     string `db:"name"`
	IsActive bool   `db:"is_active"`
}
