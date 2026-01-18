package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	db *sql.DB
}

func validatePhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}

	// Remove common formatting characters
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")

	// Check if we have at least 10 digits
	digitRegex := regexp.MustCompile(`\d`)
	digits := digitRegex.FindAllString(cleaned, -1)

	return len(digits) >= 10
}

func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{db: db}
}

func (h *Handlers) HomeHandler(c *gin.Context) {
	var workshop Workshop
	err := h.db.QueryRow(`
        SELECT id, title, description, date, location, max_capacity 
        FROM workshops 
        ORDER BY date DESC 
        LIMIT 1
    `).Scan(&workshop.ID, &workshop.Title, &workshop.Description,
		&workshop.Date, &workshop.Location, &workshop.MaxCapacity)

	if err != nil {
		c.HTML(http.StatusOK, "no_workshop.html", nil)
		return
	}

	// Count signups
	h.db.QueryRow("SELECT COUNT(*) FROM signups WHERE workshop_id = ?",
		workshop.ID).Scan(&workshop.SignupCount)

	// Check for success message
	success := c.Query("success") == "true"

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Workshop": workshop,
		"Success":  success,
	})
}

func (h *Handlers) SignupHandler(c *gin.Context) {
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"Error": "Please fill in all required fields correctly.",
		})
		return
	}

	// Get country code from form
	countryCode := c.PostForm("country_code")

	// Combine country code with phone number if phone is provided
	fullPhone := form.Phone
	if form.Phone != "" && countryCode != "" {
		fullPhone = countryCode + " " + form.Phone
	}

	// Validate phone number
	if !validatePhone(fullPhone) {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"Error": "Please enter a valid phone number with at least 7 digits.",
		})
		return
	}

	// Get workshop details for email
	var workshop Workshop
	err := h.db.QueryRow(`
        SELECT id, title, date, location 
        FROM workshops 
        WHERE id = ?
    `, form.WorkshopID).Scan(&workshop.ID, &workshop.Title, &workshop.Date, &workshop.Location)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Workshop not found"})
		return
	}

	// Insert signup with full phone number including country code
	result, err := h.db.Exec(`
        INSERT INTO signups (workshop_id, first_name, last_name, email, phone) 
        VALUES (?, ?, ?, ?, ?)
    `, form.WorkshopID, form.FirstName, form.LastName, form.Email, fullPhone)

	if err != nil {
		log.Printf("Error inserting signup: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving signup"})
		return
	}

	// Get the inserted signup ID
	signupID, _ := result.LastInsertId()

	// Create signup object for emails
	signup := Signup{
		ID:         int(signupID),
		WorkshopID: form.WorkshopID,
		FirstName:  form.FirstName,
		LastName:   form.LastName,
		Email:      form.Email,
		Phone:      fullPhone,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	// Send notification email to admin (non-blocking)
	go sendSignupNotification(signup, workshop.Title, workshop.Date)

	// Send confirmation email to participant (non-blocking)
	go sendConfirmationEmail(signup, workshop.Title, workshop.Date, workshop.Location)

	c.Redirect(http.StatusSeeOther, "/?success=true")
}

func (h *Handlers) AdminHandler(c *gin.Context) {
	// Get username from context
	username, _ := c.Get("username")

	// Check for success/error messages
	passwordChanged := c.Query("password_changed") == "true"
	passwordError := c.Query("password_error")

	// Get workshop
	var workshop Workshop
	err := h.db.QueryRow(`
        SELECT id, title, date, max_capacity 
        FROM workshops 
        ORDER BY date DESC 
        LIMIT 1
    `).Scan(&workshop.ID, &workshop.Title, &workshop.Date, &workshop.MaxCapacity)

	if err != nil {
		// No workshop exists, just show the create form
		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Workshop":        nil,
			"Signups":         []Signup{},
			"Count":           0,
			"Username":        username,
			"PasswordChanged": passwordChanged,
			"PasswordError":   passwordError,
		})
		return
	}

	// Get signups - with error logging
	rows, err := h.db.Query(`
        SELECT id, first_name, last_name, email, phone, created_at 
        FROM signups 
        WHERE workshop_id = ? 
        ORDER BY created_at DESC
    `, workshop.ID)
	if err != nil {
		// Log the actual error
		log.Printf("Error querying signups: %v", err)
		c.String(http.StatusInternalServerError, "Error loading signups: %v", err)
		return
	}
	defer rows.Close()

	var signups []Signup
	for rows.Next() {
		var s Signup
		err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Phone, &s.CreatedAt)
		if err != nil {
			log.Printf("Error scanning signup row: %v", err)
			continue
		}
		signups = append(signups, s)
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating signups: %v", err)
		c.String(http.StatusInternalServerError, "Error loading signups: %v", err)
		return
	}

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"Workshop":        workshop,
		"Signups":         signups,
		"Count":           len(signups),
		"Username":        username,
		"PasswordChanged": passwordChanged,
		"PasswordError":   passwordError,
	})
}

func (h *Handlers) CreateWorkshopHandler(c *gin.Context) {
	var form struct {
		Title        string `form:"title" binding:"required"`
		Description  string `form:"description" binding:"required"`
		WorkshopDate string `form:"workshop_date" binding:"required"`
		WorkshopTime string `form:"workshop_time" binding:"required"`
		Location     string `form:"location" binding:"required"`
		MaxCapacity  int    `form:"max_capacity" binding:"required,min=1"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the date and time
	dateTime, err := time.Parse("2006-01-02 15:04", form.WorkshopDate+" "+form.WorkshopTime)
	if err != nil {
		log.Printf("Error parsing date/time: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date or time format"})
		return
	}

	// Format as: "Saturday, March 15, 2025 at 6:00 PM"
	formattedDate := dateTime.Format("Monday, January 2, 2006 at 3:04 PM")

	_, err = h.db.Exec(`
        INSERT INTO workshops (title, description, date, location, max_capacity) 
        VALUES (?, ?, ?, ?, ?)
    `, form.Title, form.Description, formattedDate, form.Location, form.MaxCapacity)

	if err != nil {
		log.Printf("Error creating workshop: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating workshop"})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin")
}

func (h *Handlers) ChangePasswordHandler(c *gin.Context) {
	username, _ := c.Get("username")

	var form struct {
		CurrentPassword string `form:"current_password" binding:"required"`
		NewPassword     string `form:"new_password" binding:"required,min=6"`
		ConfirmPassword string `form:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBind(&form); err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=Invalid form data")
		return
	}

	// Check if new passwords match
	if form.NewPassword != form.ConfirmPassword {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=New passwords do not match")
		return
	}

	// Verify current password
	var currentHash string
	err := h.db.QueryRow("SELECT password_hash FROM admin_users WHERE username = ?", username).Scan(&currentHash)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=User not found")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(form.CurrentPassword))
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=Current password is incorrect")
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=Error updating password")
		return
	}

	// Update password
	_, err = h.db.Exec("UPDATE admin_users SET password_hash = ? WHERE username = ?", string(newHash), username)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin?password_error=Error updating password")
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin?password_changed=true")
}

func (h *Handlers) ExportCSVHandler(c *gin.Context) {
	// Get workshop
	var workshop Workshop
	err := h.db.QueryRow(`
        SELECT id, title, date 
        FROM workshops 
        ORDER BY date DESC 
        LIMIT 1
    `).Scan(&workshop.ID, &workshop.Title, &workshop.Date)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No workshop found"})
		return
	}

	// Get signups
	rows, err := h.db.Query(`
        SELECT first_name, last_name, email, phone, created_at 
        FROM signups 
        WHERE workshop_id = ? 
        ORDER BY created_at ASC
    `, workshop.ID)
	if err != nil {
		log.Printf("Error querying signups for CSV: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error loading signups"})
		return
	}
	defer rows.Close()

	// Set headers for CSV download
	filename := fmt.Sprintf("workshop-signups-%s.csv", time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"First Name", "Last Name", "Email", "Phone", "Signed Up At"})

	// Write data
	for rows.Next() {
		var firstName, lastName, email, phone, createdAt string
		err := rows.Scan(&firstName, &lastName, &email, &phone, &createdAt)
		if err != nil {
			log.Printf("Error scanning CSV row: %v", err)
			continue
		}
		writer.Write([]string{firstName, lastName, email, phone, createdAt})
	}
}
