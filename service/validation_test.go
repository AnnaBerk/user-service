package service

import (
	"errors"
	"testing"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		fio      FIO
		expected error
	}{
		{
			fio:      FIO{Name: "", Surname: ""},
			expected: errors.New("name is required"),
		},
		{
			fio:      FIO{Name: "John", Surname: ""},
			expected: errors.New("surname is required"),
		},
		{
			fio:      FIO{Name: "John", Surname: "Smith"},
			expected: nil,
		},
	}

	for _, test := range tests {
		err := test.fio.IsValid()

		if err != nil && test.expected == nil {
			t.Errorf("expected no error, but got %v", err)
		}

		if err == nil && test.expected != nil {
			t.Errorf("expected error %v, but got none", test.expected)
		}

		if err != nil && test.expected != nil && err.Error() != test.expected.Error() {
			t.Errorf("expected error %v, but got %v", test.expected, err)
		}
	}
}
