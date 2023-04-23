package postgres

import sq "github.com/Masterminds/squirrel"

func NewSelectBuilder() sq.SelectBuilder {
  return sq.SelectBuilder{}.PlaceholderFormat(sq.Dollar)
}

func NewInsertBuilder() sq.InsertBuilder {
  return sq.InsertBuilder{}.PlaceholderFormat(sq.Dollar)
}

func NewUpdateBuilder() sq.UpdateBuilder {
  return sq.UpdateBuilder{}.PlaceholderFormat(sq.Dollar)
}
