package controllers

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/revel/revel"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	lowercase    = "abcdefghijklmnopqrstuvwxyz"
	uppercase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialChars = "!@#$"
	digits       = "0123456789"
)

// Account is the controller for user authentication
type Account struct {
	*revel.Controller
}

// Index is the GET action for Login/Index page
func (c Account) Index() revel.Result {
	return c.Render()
}

// Unauthorized is the GET action for the unauthorized page
func (c Account) Unauthorized() revel.Result {
	user := &models.User{}
	_, err := c.Session.GetInto("user", user, false)
	if err != nil {
		c.Log.Error("Unauthenticated. Unable to fetch user from session. Err: %+v", err)
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	c.Flash.Out["loggedInUser"] = user.GetUserName()
	return c.Render()
}

// Login is the POST method for user login
func (c Account) Login(username, password string) revel.Result {
	c.Validation.Required(username).Message("Username is required")
	c.Validation.Required(password).Message("Password is required")

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Account.Index)
	}

	username = strings.ToLower(username)

	// Get the user details
	user, err := models.GetUserByUserName(username)
	if err != nil || user == nil {
		c.Flash.Error("Unable to process request")
		return c.Redirect(Account.Index)
	}

	// Verify the user details
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		c.Flash.Error("Invalid username and/or password")
		return c.Redirect(Account.Index)
	}

	err = c.Session.Set("user", user)
	if err != nil {
		c.Log.Errorf("Error occured while creating session. Err: %s", err.Error())
		return c.Redirect(Account.Index)
	}

	c.Flash.Out["userID"] = strconv.FormatInt(user.GetUserID(), 10)
	c.Flash.Out["loggedInUser"] = user.GetUserName()
	return c.Redirect(Home.Index)
}

// Register is the GET action for user signup page
func (c Account) Register() revel.Result {
	return c.Render()
}

// Add is the POST method for user signup
func (c Account) Add() revel.Result {
	firstName := c.Params.Get("first_name")
	lastName := c.Params.Get("last_name")
	email := c.Params.Get("email")
	username := c.Params.Get("username")
	password := c.Params.Get("password")
	confirmPassword := c.Params.Get("confirm_password")
	username = strings.ToLower(username)

	c.Validation.Required(firstName).Message("First Name is required")
	c.Validation.Required(lastName).Message("Last Name is required")
	c.Validation.Required(email).Message("Email is required")
	c.Validation.Required(username).Message("Username is required")
	c.Validation.Required(password).Message("Password is required")
	c.Validation.Required(confirmPassword).Message("Confirm Password is required")

	c.Validation.MaxSize(firstName, 50).Message("First name should be 50 characters or less")
	c.Validation.MaxSize(lastName, 50).Message("Last name should be 50 characters or less")
	c.Validation.MaxSize(email, 50).Message("Email should be 50 characters or less")
	c.Validation.MaxSize(username, 12).Message("Username should be 12 characters or less")
	c.Validation.MaxSize(password, 20).Message("Password should be maximum 20 characters")
	c.Validation.MinSize(password, 8).Message("Password must be at least 8 characters long")

	c.Validation.Match(username, regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]+$")).Message("Username should start with an alphabet and must include only alphabets (a-z, A-Z), numbers (0-9)")
	c.Validation.Match(email, regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")).Message("Email address is invalid")

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Account.Register)
	}

	if !strings.ContainsAny(password, lowercase) ||
		!strings.ContainsAny(password, uppercase) ||
		!strings.ContainsAny(password, digits) ||
		!strings.ContainsAny(password, specialChars) {
		c.Flash.Error("Invalid password. Must contain at least one - upper case letter, lowercase letter, digit, and special character (!@#$)")
		return c.Redirect(Account.Register)
	}

	if password != confirmPassword {
		c.Flash.Error("Passwords do not match")
		return c.Redirect(Account.Register)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Flash.Error("Unable to add the user at the moment")
		return c.Redirect(Account.Register)
	}

	user := &models.User{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Username:     username,
		Password:     string(hashedPassword),
		MaxItemCount: common.UserMaxAllowedItems,
		MaxItemSpace: common.UserMaxAllowedSpace,
	}

	added, err := user.Add()
	if err != nil ||
		!added {
		log.Printf("Error while adding user. Err: %+v", err)
		c.Flash.Error("Unable to add user at the moment. Please try after some time.")
		return c.Redirect(Account.Register)
	}

	c.Flash.Success("Please login to continue")
	return c.Redirect(Account.Index)
}

// Logout clears all the application sessions
func (c Account) Logout() revel.Result {
	for i := range c.Session {
		delete(c.Session, i)
	}
	return c.Redirect(Account.Index)
}
