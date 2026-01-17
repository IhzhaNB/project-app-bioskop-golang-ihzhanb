package entity

type Genre struct {
	BaseSimple
	Name string `db:"name"`
}
