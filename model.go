package main

import (
	"errors"
	"net/http"
)

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     string   `json:"age"`
	Friends []string `json:"friends"`
}

func (u User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (u User) Bind(r *http.Request) error {
	if u.Name == "" {
		return errors.New("Отсутсвует необходимое поле имени пользователя")
	}
	if u.Age == "" {
		return errors.New("Отсутсвует необходимое поле возраста пользователя")
	}
	return nil
}

type Friends struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

func (f Friends) Bind(r *http.Request) error {
	if f.SourceID == "" || f.TargetID == "" {
		return errors.New("Отсутсвует необходимое поле ID друга")
	}
	return nil
}

type TargetID struct {
	TargetID string `json:"target_id"`
}

func (t TargetID) Bind(r *http.Request) error {
	if t.TargetID == "" {
		return errors.New("Отсутсвует необходимое поле ID пользователя")
	}
	return nil
}

type NewAge struct {
	NewAge string `json:"new_age"`
}

func (n NewAge) Bind(r *http.Request) error {
	if n.NewAge == "" {
		return errors.New("Отсутсвует необходимое поле нового возраста")
	}
	return nil
}
