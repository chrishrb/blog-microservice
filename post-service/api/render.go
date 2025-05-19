package api

import "net/http"

func (c PostCreate) Bind(r *http.Request) error {
	return nil
}

func (c PostUpdate) Bind(r *http.Request) error {
	return nil
}

func (c Post) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}


func (c CommentCreate) Bind(r *http.Request) error {
	return nil
}

func (c CommentUpdate) Bind(r *http.Request) error {
	return nil
}

func (c Comment) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
