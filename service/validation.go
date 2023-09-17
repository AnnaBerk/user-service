package service

import "errors"

func (f *FIO) IsValid() error {
	if f.Name == "" {
		return errors.New("name is required")
	}
	if f.Surname == "" {
		return errors.New("surname is required")
	}
	// остальная валидация во многом зависит от бизнес-логики и требований к данным,
	//можно проверить длину параметров, что они содержат только буквы и т.д.

	return nil
}
