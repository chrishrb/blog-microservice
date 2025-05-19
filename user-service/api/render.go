package api

import "net/http"

func (c UserCreate) Bind(r *http.Request) error {
	return nil
}

func (c UserUpdate) Bind(r *http.Request) error {
	return nil
}

func (c User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}


func (c LoginRequest) Bind(r *http.Request) error {
	return nil
}

func (c AuthResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

