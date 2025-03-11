package main

import (
	"advanced-algorithms/virtual-keyboard/postgresql"
	"context"
	"log/slog"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StartResponse struct {
	Hash         uuid.UUID   `json:"hash"`
	Combinations [][]int32 `json:"combinations"`
}

type CheckRequest struct {
	Hash         *string      `json:"hash"`
	Combinations *[][]int32 `json:"combinations"`
}

type UserReponse struct {
	Username string `json:"username"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type User struct {
	Username string   `json:"username"`
	Password []int32 `json:",omit"`
}

var possible_combinations = [][][]int32{}

var curr_index = 0

var login_attempts = make(map[uuid.UUID]*[][]int32)

var users = []User{
	{
		Username: "Among",
		Password: []int32{1, 2, 3, 4, 5},
	},
}

func StartLogin(c fuego.ContextNoBody) (*StartResponse, error) {
	options := possible_combinations[curr_index]

	if curr_index < len(possible_combinations)-1 {
		curr_index += 1
	} else {
		curr_index = 0
	}

	hash := uuid.New()
	login_attempts[hash] = &options

	response := StartResponse{
		Hash:         hash,
		Combinations: options,
	}

	return &response, nil
}

func CheckLogin(c fuego.ContextWithBody[CheckRequest]) (*UserReponse, error) {
	body, err := c.Body()

	if err != nil {
		return nil, fuego.BadRequestError{}
	}

	hash_uuid, err := uuid.Parse(*body.Hash)

	if err != nil {
		return nil, fuego.BadRequestError{Detail: "Hash is not valid UUID"}
	}

	session := login_attempts[hash_uuid]

	if session == nil {
		return nil, fuego.NotFoundError{Detail: "Session with this hash does not exist"}
	}

	delete(login_attempts, hash_uuid)

	for _, request_combination := range *body.Combinations {
		valid := false
		for _, possible := range *session {
			if request_combination[0] == possible[0] && request_combination[1] == possible[1] {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fuego.BadRequestError{Detail: "Hash did not match this combination"}
		}
	}

	for _, user := range users {
		valid := true
		for password_digit_i, password_digit := range user.Password {
			if (*body.Combinations)[password_digit_i][0] != password_digit && (*body.Combinations)[password_digit_i][1] != password_digit {
				valid = false
				break
			}
		}

		if valid {
			response := UserReponse{
				Username: user.Username,
			}

			return &response, nil
		}
	}

	return nil, fuego.UnauthorizedError{Detail: "Wrong password"}
}

var schema = `
create table combinations (
    options int[][]
);

create table users (
    username varchar(50) primary key,
    password int[]
);
`

func MigrateDB(db *pgxpool.Conn) error {
	_, err := db.Exec(context.Background(), schema)
	slog.Info("Migrating database...")

	if err != nil && !strings.Contains(err.Error(), "42P07") {
		slog.Error("Failed to migrate database", "error", err)
		return err
	}

	slog.Info("Database migrated")

	return nil
}

func main() {
	s := fuego.NewServer(
		fuego.WithAddr("localhost:8080"),
	)

	fuego.Options(s, "/login", StartLogin)
	fuego.Post(s, "/login", CheckLogin)

	postgresql.Connect()
	defer postgresql.Close()

	conn, err := postgresql.GetConection()

	if err != nil {
		panic(err)
	}

	err = MigrateDB(conn)

	if err != nil {
		panic(err)
	}

	queries := postgresql.New(conn)

	combinations, err := queries.GetCombinations(context.Background())

	if err != nil {
		slog.Error("Failed to get combinations", "error", err)
		panic(err)
	}

	possible_combinations = combinations

	err = s.Run()

	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
