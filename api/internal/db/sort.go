package db

import "strings"

type SortDirection string

const (
	SortAscending  SortDirection = "asc"
	SortDescending SortDirection = "desc"
)

type Sort struct {
	Field     string
	Direction SortDirection
}

func AscSort(field string) Sort {
	return Sort{Field: field, Direction: SortAscending}
}

func DescSort(field string) Sort {
	return Sort{Field: field, Direction: SortDescending}
}

func (s Sort) IsZero() bool {
	return strings.TrimSpace(s.Field) == ""
}

func (s Sort) String() string {
	field := strings.TrimSpace(s.Field)
	if field == "" {
		return ""
	}
	if s.Direction == SortDescending {
		return "-" + field
	}
	return field
}
