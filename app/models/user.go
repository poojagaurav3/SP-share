package models

import (
	"fmt"
	"time"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/database"
)

// User is the struct for application user
type User struct {
	tableName      struct{}  `sql:"AppUser,alias:usr"`
	UserID         int64     `sql:"user_id,pk"`
	FirstName      string    `sql:"first_name"`
	LastName       string    `sql:"last_name"`
	Email          string    `sql:"email"`
	Username       string    `sql:"username"`
	Password       string    `sql:"password"`
	CreationTime   time.Time `sql:"creation_time"`
	LastUpdated    time.Time `sql:"last_updated"`
	IsAdmin        bool      `sql:"is_admin,default:false"`
	WorkflowStatus int       `sql:"workflow_status"`
	MaxItemCount   int       `sql:"max_item_count"`
	MaxItemSpace   float32   `sql:"max_item_space"`
}

// UserKeyVal holds the key-value pair for user model
type UserKeyVal struct {
	tableName struct{} `sql:"AppUser"`
	UserID    int64    `sql:"user_id,pk"`
	FirstName string   `sql:"first_name"`
	LastName  string   `sql:"last_name"`
}

// UserLimits is the view model for setting user level limits
type UserLimits struct {
	UserID       int64
	Users        []*UserKeyVal
	MaxItemCount int
	MaxItemSpace float32
}

// GetUserID returns the user id of the user model
func (model *User) GetUserID() int64 {
	if model == nil {
		return -1
	}
	return model.UserID
}

// GetUserName returns the name of the logged-in user
func (model *User) GetUserName() string {
	if model == nil {
		return ""
	}
	return fmt.Sprintf("%s %s", model.FirstName, model.LastName)
}

// Add inserts a user to database
func (model *User) Add() (bool, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return false, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	// Insert user model into database
	res, err := client.GetPGClient().Model(model).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return false, fmt.Errorf("Unable to insert user into database. Err: %s", err.Error())
	}

	return (res.RowsAffected() == 1), nil
}

// UpdateLimits updates the item upload limits associated with a user
func (model *User) UpdateLimits(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update user limits")
	}

	// Update user model in database
	res, err := client.GetPGClient().Model(model).
		Column("max_item_count").
		Column("max_item_space").
		WherePK().
		Update()
	if err != nil {
		log.Errorf("Unable to update user into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to update user limits")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to update user limits")
	}

	return nil
}

// GetUserByCredentials get a user from database basis the username and password
func GetUserByCredentials(username, password string) (*User, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	user := &User{}
	client.GetPGClient().Model(user).Where("username = ?", username).Where("password = ?", password).Select()

	// Check for empty user
	if *user == (User{}) {
		return nil, nil
	}

	return user, nil
}

// GetUserByUserID get a user from database basis the userID
func GetUserByUserID(userID int64) (*User, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	user := &User{
		UserID: userID,
	}
	err = client.GetPGClient().Select(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUserName get a user from database basis the userID
func GetUserByUserName(username string) (*User, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	user := &User{}
	err = client.GetPGClient().Model(user).Where("username = ?", username).Select()
	if err != nil {
		return nil, err
	}

	if user.UserID < 1 {
		return nil, fmt.Errorf("Invalid username")
	}

	return user, nil
}

// GetAllUsersPendingApproval returns list of all the users that are pending for admin approval
func GetAllUsersPendingApproval(userID int64) ([]*User, error) {
	var users []*User

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	// Get the User
	userModel, err := GetUserByUserID(userID)
	if err != nil {
		return nil, err
	}

	if !userModel.IsAdmin {
		return nil, fmt.Errorf("unauthorized - only admins can access the content")
	}

	actionID := common.WorkflowStatusPending.GetStatusID()
	// Get all the groups
	client.GetPGClient().Model(&users).Where("workflow_status = ?", actionID).Select()

	return users, nil
}

// ApproveOrReject is used by admin to approve/reject a group creation
func (model *User) ApproveOrReject(log logger.MultiLogger, approve bool) error {
	// Note: Authz check is already done at this point

	var actionID common.WorkflowStatus

	if approve {
		actionID = common.WorkflowStatusApproved
	} else {
		actionID = common.WorkflowStatusRejected
	}

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update the group at the moment")
	}

	model.WorkflowStatus = actionID.GetStatusID()

	res, err := client.GetPGClient().Model(model).WherePK().Set("workflow_status = ?", model.WorkflowStatus).Update()
	if err != nil {
		log.Errorf("Unable to update the user model. Error: %s", err.Error())
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	return nil
}

// GetAllUsersKeyVal returns list of all approved non-admin users
func GetAllUsersKeyVal(log logger.MultiLogger) ([]*UserKeyVal, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	var users []*UserKeyVal

	err = client.GetPGClient().Model(&users).
		Where("workflow_status = ?", common.WorkflowStatusApproved.GetStatusID()).
		Where("is_admin = ?", false).
		Order("first_name ASC").
		Order("last_name ASC").
		Select()
	if err != nil {
		log.Errorf("Unable to get the list of users from database. Error: %s", err.Error())
		return nil, err
	}

	return users, nil
}
