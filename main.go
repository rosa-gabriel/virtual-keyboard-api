package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type StartResponse struct {
	Hash    uuid.UUID    `json:"hash"`
	Options [5][2]uint32 `json:"options"`
}

type CheckRequest struct {
	Hash    *string       `json:"hash"`
	Options *[5][2]uint32 `json:"options"`
}

type UserReponse struct {
	Username string `json:"username"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type User struct {
	Username string    `json:"username"`
	Password [5]uint32 `json:",omit"`
}

var possible_combinations = [][5][2]uint32{
	{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}},
	{{1, 2}, {3, 5}, {4, 6}, {7, 8}, {9, 10}},
	{{1, 2}, {3, 4}, {5, 7}, {6, 8}, {9, 10}},
	{{5, 2}, {3, 4}, {1, 6}, {7, 8}, {9, 10}},
	{{1, 2}, {3, 4}, {5, 9}, {7, 8}, {6, 10}},
}

var curr_index = 0

var login_attempts = make(map[uuid.UUID]*[5][2]uint32)

var users = []User{
	{
		Username: "Among",
		Password: [5]uint32{1, 2, 3, 4, 5},
	},
}

func StartLogin(w http.ResponseWriter, r *http.Request) {
	options := possible_combinations[curr_index]

	if curr_index < len(possible_combinations)-1 {
		curr_index += 1
	} else {
		curr_index = 0
	}

	hash := uuid.New()
	login_attempts[hash] = &options

	response := StartResponse{
		Hash:    hash,
		Options: options,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CheckLogin(w http.ResponseWriter, r *http.Request) {
	body_bytes, err := io.ReadAll(r.Body)

	var body CheckRequest

	err = json.Unmarshal(body_bytes, &body)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
        w.WriteHeader(400)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: "Body is not in json format",
		})
		return
	}

	if body.Hash == nil || body.Options == nil {
        w.WriteHeader(400)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: "Missing required body parts",
		})
		return
	}

	hash_uuid, err := uuid.Parse(*body.Hash)

	if err != nil {
        slog.Info("error", "error", err)
        w.WriteHeader(400)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: "Hash is not valid UUID",
		})
		return
	}

	session := login_attempts[hash_uuid]

	if session == nil {
        w.WriteHeader(404)
		json.NewEncoder(w).Encode(ErrorResponse{
			Message: "Session with this hash does not exist",
		})
		return
	}

	delete(login_attempts, hash_uuid)

	for _, request_combination := range body.Options {
		valid := false
		for _, possible := range session {
			if request_combination[0] == possible[0] && request_combination[1] == possible[1] {
				valid = true
				break
			}
		}

		if !valid {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(ErrorResponse{
				Message: "Hash did not match this combination",
			})
			return
		}
	}

	for _, user := range users {
		valid := true
		for password_digit_i, password_digit := range user.Password {
			if body.Options[password_digit_i][0] != password_digit && body.Options[password_digit_i][1] != password_digit {
				valid = false
				break
			}
		}

		if valid {
			response := UserReponse{
				Username: user.Username,
			}

			json.NewEncoder(w).Encode(response)
            return
		}
	}

	w.WriteHeader(401)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: "Wrong password",
	})
	return
}
func main() {
	http.HandleFunc("OPTIONS /login", StartLogin)
	http.HandleFunc("POST /login", CheckLogin)

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
