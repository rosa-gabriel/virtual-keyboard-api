package main

import (
	"advanced-algorithms/virtual-keyboard/postgresql"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type StartResponse struct {
	Hash         uuid.UUID `json:"hash"`
	Combinations [][]int32 `json:"combinations"`
}

type CheckRequest struct {
	Hash         *string      `json:"hash"`
	Combinations *[4][2]int32 `json:"combinations"`
}

type UserReponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var possible_combinations = [][][]int32{}
var switch_combinations [][]int32

var curr_index = 0

var login_attempts = make(map[uuid.UUID]*[][]int32)

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

func GenerateSwitchCombinations(n int) [][]int32 {
	totalCombinations := 1 << n

	result := make([][]int32, totalCombinations)

	for i := 0; i < totalCombinations; i++ {
		combination := make([]int32, n)
		for j := 0; j < n; j++ {
			if i&(1<<j) != 0 {
				combination[j] = 1
			} else {
				combination[j] = 0
			}
		}
		result[i] = combination
	}
	return result
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

	conn, err := postgresql.GetConection()

	if err != nil {
		panic(err)
	}

	queries := postgresql.New(conn)

	users, err := queries.GetUsers(context.Background())

	if err != nil {
		slog.Error("Failed to get users")
		panic(err)
	}

	user_combinations := *body.Combinations

	for _, s_c := range switch_combinations {
		password := fmt.Sprintf("%d%d%d%d", user_combinations[0][s_c[0]], user_combinations[1][s_c[1]], user_combinations[2][s_c[2]], user_combinations[3][s_c[3]])

		for _, user := range users {
			if VerifyPassword(password, user.Password.String) {
				response := UserReponse{
					Username: user.Username,
					Email:    user.Email.String,
				}

				return &response, nil
			}
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
    email varchar(50),
    password varchar(100) unique
);

INSERT INTO public.combinations ("options") VALUES
	 ('{{1,2},{3,4},{5,7},{6,8},{9,10}}'),
	 ('{{5,2},{3,4},{1,6},{7,8},{9,10}}'),
	 ('{{1,2},{3,4},{5,9},{7,8},{6,10}}'),
	 ('{{1,2},{3,4},{5,6},{7,8},{9,10}}'),
	 ('{{1,2},{3,5},{4,6},{7,8},{9,10}}');

INSERT INTO public.users (username,email,"password") VALUES
	 ('Among Us','among@us.br','$2a$14$Mzu9bIkdj.NSA6vNjLHl.O4xo5.AagKMnQvgwbhhnSLQ48ji/Dfry'),
	 ('Paulo Sérgio','paulo@kuba.ch','$2a$14$KMqaccLtCYFncwTgMkQ6FueEtoufOkUPmPNQ03UnC2G0fcejzVN2u');
`

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 2)
	return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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

	switch_combinations = GenerateSwitchCombinations(4)

	err = s.Run()

	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
