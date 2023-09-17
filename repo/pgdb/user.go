package pgdb

import (
	"context"
	"fmt"
	"user-service/api_clients/model"
	"user-service/pkg/psql"
)

type UserRepo struct {
	db *psql.Postgres
}

func NewUserRepo(pg *psql.Postgres) *UserRepo {
	return &UserRepo{
		db: pg,
	}
}

type User struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

func (ur *UserRepo) Save(ctx context.Context, user model.User) error {
	query := `
		INSERT INTO users (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := ur.db.Pool.QueryRow(ctx, query, user.Name, user.Surname, user.Patronymic, user.Age, user.Gender, user.Nationality).Scan(&id)
	if err != nil {
		return err
	}

	user.ID = id

	return nil
}

func (r *UserRepo) GetUsers(page, size int, filter string) ([]model.User, error) {
	query := `
	SELECT id, name, surname, patronymic 
	FROM users 
	WHERE name LIKE $1 
	LIMIT $2 OFFSET $3`

	filter = "%" + filter + "%"
	offset := (page - 1) * size

	rows, err := r.db.Pool.Query(context.Background(), query, filter, size, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err = rows.Scan(&user.ID, &user.Name, &user.Surname, &user.Patronymic)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepo) AddUser(user model.User) (int, error) {
	query := `
	INSERT INTO users (name, surname, patronymic)
	VALUES ($1, $2, $3)
	RETURNING id`

	var id int
	err := r.db.Pool.QueryRow(context.Background(), query, user.Name, user.Surname, user.Patronymic).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *UserRepo) DeleteUser(id int) error {
	query := `
	DELETE FROM users 
	WHERE id = $1`

	_, err := r.db.Pool.Exec(context.Background(), query, id)
	return err
}

func (r *UserRepo) UpdateUser(user model.User) error {
	query := `
	UPDATE users 
	SET name = $1, surname = $2, patronymic = $3 
	WHERE id = $4`

	result, err := r.db.Pool.Exec(context.Background(), query, user.Name, user.Surname, user.Patronymic, user.ID)
	if err != nil {
		return err
	}
	// Проверка количества обновленных строк
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("User with ID %d not found", user.ID)
	}
	return nil
}
