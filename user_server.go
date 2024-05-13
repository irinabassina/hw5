package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"net/http"
)

type userServer struct {
	us  *userService
	mux *chi.Mux
}

func newUserServer(dbName string) *userServer {
	us := newUserService(dbName)
	r := chi.NewRouter()

	uServer := &userServer{us: us, mux: r}

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Server is up and running!"))
		})
		r.Post("/create", uServer.CreateUser)
		r.Get("/{id}", uServer.GetUser)
		r.Post("/make_friends", uServer.MakeFriends)
		r.Delete("/user", uServer.DeleteUser)
		r.Get("/friends/{id}", uServer.GetFriends)
		r.Put("/{id}", uServer.UpdateAge)
	})

	return uServer
}

func (uServer *userServer) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := &User{}
	if err := render.Bind(r, user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := uServer.us.storeUser(user)
	render.Status(r, http.StatusCreated)
	render.PlainText(w, r, id)
}

func (uServer *userServer) UpdateAge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	newAge := &NewAge{}
	if err := render.Bind(r, newAge); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !uServer.us.updateAge(id, newAge.NewAge) {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	render.PlainText(w, r, "Возраст пользователя успешно обновлён")
}

func (uServer *userServer) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user := uServer.us.getUser(id)
	if user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	render.Render(w, r, user)
}

func (uServer *userServer) GetFriends(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	friends, ok := uServer.us.getFriends(id)
	if !ok {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	render.JSON(w, r, friends)
}

func (uServer *userServer) DeleteUser(w http.ResponseWriter, r *http.Request) {
	target := &TargetID{}
	if err := render.Bind(r, target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name, ok := uServer.us.deleteUser(target.TargetID)
	if !ok {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	render.PlainText(w, r, name)
}

func (uServer *userServer) MakeFriends(w http.ResponseWriter, r *http.Request) {
	friends := &Friends{}
	if err := render.Bind(r, friends); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sName, tName, ok := uServer.us.makeFriends(friends)
	if !ok {
		http.Error(w, "Друг не найден по id", http.StatusBadRequest)
		return
	}
	render.PlainText(w, r, fmt.Sprintf("%s и %s теперь друзья", sName, tName))
}

func (uServer *userServer) dropDB() {
	uServer.us.dropDB()
}
