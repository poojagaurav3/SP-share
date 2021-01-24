package controllers

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/revel/revel"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/models"
)

// Group is the controller for user group related functions
type Group struct {
	*revel.Controller
}

// Index is the GET action for Group/Index page
func (c Group) Index() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	// Get all the groups applicable to user
	groups, err := models.GetAllGroups(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch list of groups. Error: %s", err.Error())
	}

	return c.Render(groups)
}

// Create is the action method for creating a group
func (c Group) Create(group string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}
	c.Validation.Required(group).Message("Group name is required")
	c.Validation.MaxSize(group, 60).Message("Group name should be less than 60 characters")
	c.Validation.Match(group, regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_.' ]+$")).Message("Group name should start with an alphabet and must include only alphabets (a-z, A-Z), numbers (0-9) and symbols (. and _)")
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Group.Index)
	}

	groupModel := &models.Group{
		GroupName:    group,
		CreatedBy:    intUserID,
		MaxItemCount: common.GroupMaxAllowedItems,
		MaxItemSpace: common.GroupMaxAllowedSpace,
	}

	// Push the group into database table
	err = groupModel.Add(c.Log)
	if err != nil {
		c.Flash.Error("Unable to add group at the moment. Please try after sometime.")
		return c.Redirect(Group.Index)
	}

	c.Flash.Success("Group created successfully")
	return c.Redirect(Group.Index)
}

// Details displays the details of a group ID
func (c Group) Details(id int64) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	group, err := models.GetGroupDetails(c.Log, intUserID, id)
	if err != nil {
		c.Log.Errorf("Unable to fetch group details. Error: %s", err.Error())
		c.Flash.Error("Internal error. Please try after sometime.")
	}

	return c.Render(group)
}

// MapUser maps a user with the given username to the group
func (c Group) MapUser(username, groupID string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	username = strings.ToLower(username)

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	intGroupID, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid Group ID found - %s. Error: %s", groupID, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Group.Index)
	}

	c.Validation.Required(username).Message("Username is required")
	c.Validation.MaxSize(username, 12).Message("Username should be 12 characters or less")
	c.Validation.Match(username, regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_.]+$")).Message("Item name should start with an alphabet and must include only alphabets (a-z, A-Z), numbers (0-9) and symbols (. and _)")

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect("/groups/%d", intGroupID)
	}

	user, err := models.GetUserByUserName(username)
	if err != nil {
		c.Log.Errorf("Invalid username - %s. Error: %+v", username, err)
		c.Flash.Error("Invalid username provided")
		return c.Redirect("/groups/%d", intGroupID)
	}

	userGroupMap := &models.UserGroupMap{
		UserID:    user.UserID,
		GroupID:   intGroupID,
		CreatedBy: intUserID,
	}

	err = userGroupMap.Add(c.Log, intUserID)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect("/groups/%d", intGroupID)
	}

	c.Flash.Success("Successfully tagged the user to the group")
	return c.Redirect("/groups/%d", intGroupID)
}

// RequestLeadAccess creates a new request for a member to be promoted to group lead
func (c Group) RequestLeadAccess(groupID string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	intGroupID, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid Group ID found - %s. Error: %s", groupID, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Group.Index)
	}

	userGroupMap := &models.UserGroupMap{
		UserID:  intUserID,
		GroupID: intGroupID,
	}

	err = userGroupMap.UpgradeToGroupLead(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Group.Index)
	}

	c.Flash.Success("Successfully created a request for group leader access")
	return c.Redirect(Group.Index)
}
