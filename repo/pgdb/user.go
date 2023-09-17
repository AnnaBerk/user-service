package pgdb

import (
	"context"
	"fmt"
	"strings"
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
	fields := []string{}
	values := []interface{}{}
	placeholders := []string{}

	// Здесь мы добавляем обязательные поля
	fields = append(fields, "name", "surname")
	values = append(values, user.Name, user.Surname)

	// Динамическое добавление полей
	if user.Patronymic != "" {
		fields = append(fields, "patronymic")
		values = append(values, user.Patronymic)
	}

	if user.Age != 0 {
		fields = append(fields, "age")
		values = append(values, user.Age)
	}

	if user.Gender != "" {
		fields = append(fields, "gender")
		values = append(values, user.Gender)
	}

	if user.Nationality != "" {
		fields = append(fields, "nationality")
		values = append(values, user.Nationality)
	}

	// Генерация плейсхолдеров для SQL-запроса
	for i := range values {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}

	query := fmt.Sprintf(`
        INSERT INTO users (%s)
        VALUES (%s)
        RETURNING id`,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	var id int
	err := r.db.Pool.QueryRow(context.Background(), query, values...).Scan(&id)
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
	var queryBuilder strings.Builder
	queryBuilder.WriteString("UPDATE users SET ")

	// Список для значений
	args := []interface{}{}
	argNumber := 1

	if user.Name != "" {
		queryBuilder.WriteString(fmt.Sprintf("name = $%d,", argNumber))
		args = append(args, user.Name)
		argNumber++
	}

	if user.Surname != "" {
		queryBuilder.WriteString(fmt.Sprintf("surname = $%d,", argNumber))
		args = append(args, user.Surname)
		argNumber++
	}

	if user.Patronymic != "" {
		queryBuilder.WriteString(fmt.Sprintf("patronymic = $%d,", argNumber))
		args = append(args, user.Patronymic)
		argNumber++
	}

	if user.Age != 0 {
		queryBuilder.WriteString(fmt.Sprintf("age = $%d,", argNumber))
		args = append(args, user.Age)
		argNumber++
	}

	if user.Gender != "" {
		queryBuilder.WriteString(fmt.Sprintf("gender = $%d,", argNumber))
		args = append(args, user.Gender)
		argNumber++
	}

	if user.Nationality != "" {
		queryBuilder.WriteString(fmt.Sprintf("nationality = $%d,", argNumber))
		args = append(args, user.Nationality)
		argNumber++
	}

	query := queryBuilder.String()
	query = query[:len(query)-1]

	query += fmt.Sprintf(" WHERE id = $%d", argNumber)
	args = append(args, user.ID)

	result, err := r.db.Pool.Exec(context.Background(), query, args...)
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
