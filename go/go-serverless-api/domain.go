package main

// Todo represents an action that has to be done in the future
type Todo struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
