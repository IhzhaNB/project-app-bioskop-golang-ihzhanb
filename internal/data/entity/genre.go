package entity

type Genre struct {
	BaseNoDelete
	Name string `db:"name"`
}
