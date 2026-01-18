package main

type Workshop struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	SignupCount int    `json:"signup_count"`
	MaxCapacity int    `json:"max_capacity"`
}

type Signup struct {
	ID         int    `json:"id"`
	WorkshopID int    `json:"workshop_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email" binding:"required,email"`
	Phone      string `json:"phone"`
	CreatedAt  string `json:"created_at"`
}

type SignupForm struct {
	WorkshopID int    `form:"workshop_id" binding:"required"`
	FirstName  string `form:"first_name" binding:"required"`
	LastName   string `form:"last_name" binding:"required"`
	Email      string `form:"email" binding:"required,email"`
	Phone      string `form:"phone" binding:"omitempty,min=10"`
}
