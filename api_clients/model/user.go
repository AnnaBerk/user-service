package model

type User struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Surname     string `json:"surname" db:"surname"`
	Patronymic  string `json:"patronymic" db:"patronymic"`
	Age         int    `json:"age" db:"age"`
	Gender      string `json:"gender" db:"gender"`
	Nationality string `json:"nationality" db:"nationality"`
}
