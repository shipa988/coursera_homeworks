package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		srv.handleProfile(w, r)
	case "/user/create":
		if r.Method == "POST" {
			if r.Header.Get("X-Auth") == "100500" {
				srv.handleCreate(w, r)
			} else {
				answer, answerErr := json.Marshal(MyResponse{Error: "unauthorized", Response: nil})
				if answerErr != nil {
					w.WriteHeader(500)
				} else {
					w.WriteHeader(403)
					w.Write(answer)
				}
			}
		} else {
			answer, answerErr := json.Marshal(MyResponse{Error: "bad method", Response: nil})
			if answerErr != nil {
				w.WriteHeader(500)
			} else {
				w.WriteHeader(406)
				w.Write(answer)
			}
		}
	default:
		answer, answerErr := json.Marshal(MyResponse{Error: "unknown method", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(404)
			w.Write(answer)
		}
	}
}

func (srv *MyApi) handleProfile(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	r.ParseForm()
	var answer []byte
	var answerErr error
	_ProfileParams := ProfileParams{}

	//Login

	login := r.FormValue("login")

	if login == "" {
		answer, answerErr = json.Marshal(MyResponse{Error: "login must me not empty", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}
	_ProfileParams.Login = login
	_user, _error := srv.Profile(ctx, _ProfileParams)

	if _error != nil {
		switch _error.(type) {
		case ApiError:
			w.WriteHeader(_error.(ApiError).HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		answer, answerErr = json.Marshal(MyResponse{Error: _error.Error(), Response: nil})
	} else {
		answer, answerErr = json.Marshal(MyResponse{Error: "", Response: _user})
	}
	if answerErr != nil {
		w.WriteHeader(500)
	} else {
		w.Write(answer)
	}
	return

}

func (srv *MyApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	r.ParseForm()
	var answer []byte
	var answerErr error
	_CreateParams := CreateParams{}

	//Login

	login := r.FormValue("login")

	if login == "" {
		answer, answerErr = json.Marshal(MyResponse{Error: "login must me not empty", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	loginlen := len(login)

	if loginlen < 10 {

		answer, answerErr = json.Marshal(MyResponse{Error: "login len must be >= 10", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_CreateParams.Login = login

	//Name
	name := r.FormValue("full_name")

	_CreateParams.Name = name

	//Status

	status := r.FormValue("status")

	if status == "" {
		status = "user"
	}

	if status != "user" && status != "moderator" && status != "admin" {
		answer, answerErr = json.Marshal(MyResponse{Error: "status must be one of " + strings.ReplaceAll("[user moderator admin]", " ", ", "), Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_CreateParams.Status = status

	//Age

	age := r.FormValue("age")

	agelen, err := strconv.Atoi(age)
	if err != nil {
		answer, answerErr = json.Marshal(MyResponse{Error: "age must be int", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	if agelen < 0 {

		answer, answerErr = json.Marshal(MyResponse{Error: "age must be >= 0", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	if agelen > 128 {

		answer, answerErr = json.Marshal(MyResponse{Error: "age must be <= 128", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_CreateParams.Age = agelen
	_newuser, _error := srv.Create(ctx, _CreateParams)

	if _error != nil {
		switch _error.(type) {
		case ApiError:
			w.WriteHeader(_error.(ApiError).HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		answer, answerErr = json.Marshal(MyResponse{Error: _error.Error(), Response: nil})
	} else {
		answer, answerErr = json.Marshal(MyResponse{Error: "", Response: _newuser})
	}
	if answerErr != nil {
		w.WriteHeader(500)
	} else {
		w.Write(answer)
	}
	return

}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		if r.Method == "POST" {
			if r.Header.Get("X-Auth") == "100500" {
				srv.handleCreate(w, r)
			} else {
				answer, answerErr := json.Marshal(MyResponse{Error: "unauthorized", Response: nil})
				if answerErr != nil {
					w.WriteHeader(500)
				} else {
					w.WriteHeader(403)
					w.Write(answer)
				}
			}
		} else {
			answer, answerErr := json.Marshal(MyResponse{Error: "bad method", Response: nil})
			if answerErr != nil {
				w.WriteHeader(500)
			} else {
				w.WriteHeader(406)
				w.Write(answer)
			}
		}
	default:
		answer, answerErr := json.Marshal(MyResponse{Error: "unknown method", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(404)
			w.Write(answer)
		}
	}
}

func (srv *OtherApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	r.ParseForm()
	var answer []byte
	var answerErr error
	_OtherCreateParams := OtherCreateParams{}

	//Username

	username := r.FormValue("username")

	if username == "" {
		answer, answerErr = json.Marshal(MyResponse{Error: "username must me not empty", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	usernamelen := len(username)

	if usernamelen < 3 {

		answer, answerErr = json.Marshal(MyResponse{Error: "username len must be >= 3", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_OtherCreateParams.Username = username

	//Name
	name := r.FormValue("account_name")

	_OtherCreateParams.Name = name

	//Class

	class := r.FormValue("class")

	if class == "" {
		class = "warrior"
	}

	if class != "warrior" && class != "sorcerer" && class != "rouge" {
		answer, answerErr = json.Marshal(MyResponse{Error: "class must be one of " + strings.ReplaceAll("[warrior sorcerer rouge]", " ", ", "), Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_OtherCreateParams.Class = class

	//Level

	level := r.FormValue("level")

	levellen, err := strconv.Atoi(level)
	if err != nil {
		answer, answerErr = json.Marshal(MyResponse{Error: "level must be int", Response: nil})
		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	if levellen < 1 {

		answer, answerErr = json.Marshal(MyResponse{Error: "level must be >= 1", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	if levellen > 50 {

		answer, answerErr = json.Marshal(MyResponse{Error: "level must be <= 50", Response: nil})

		if answerErr != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(400)
			w.Write(answer)
		}
		return
	}

	_OtherCreateParams.Level = levellen
	_otheruser, _error := srv.Create(ctx, _OtherCreateParams)

	if _error != nil {
		switch _error.(type) {
		case ApiError:
			w.WriteHeader(_error.(ApiError).HTTPStatus)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		answer, answerErr = json.Marshal(MyResponse{Error: _error.Error(), Response: nil})
	} else {
		answer, answerErr = json.Marshal(MyResponse{Error: "", Response: _otheruser})
	}
	if answerErr != nil {
		w.WriteHeader(500)
	} else {
		w.Write(answer)
	}
	return

}
