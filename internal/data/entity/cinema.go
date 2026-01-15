package entity

type Cinema struct {
	Base
	Name     string `db:"name"`
	Location string `db:"location"`
	City     string `db:"city"`
}
