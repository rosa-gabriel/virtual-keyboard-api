// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package postgresql

import (
	"context"
)

const getCombinations = `-- name: GetCombinations :many
SELECT options FROM combinations
`

func (q *Queries) GetCombinations(ctx context.Context) ([][][]int32, error) {
	rows, err := q.db.Query(ctx, getCombinations)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items [][][]int32
	for rows.Next() {
		var options [][]int32
		if err := rows.Scan(&options); err != nil {
			return nil, err
		}
		items = append(items, options)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsers = `-- name: GetUsers :many
SELECT username, email, password FROM users
`

func (q *Queries) GetUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.Query(ctx, getUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.Username, &i.Email, &i.Password); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
