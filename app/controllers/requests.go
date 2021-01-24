package controllers

import (
	"strconv"

	"github.com/revel/revel"
	"github.com/sp-share/app/models"
)

// Requests is the controller for admin approval request workflow
type Requests struct {
	*revel.Controller
}

// Groups is the GET action for getting groups pending approval
func (c Requests) Groups() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	// Authorize
	user, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if user == nil {
		c.Log.Errorf("Unable to fetch user details from database")
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get all the groups applicable to user
	groups, err := models.GetAllGroupsPendingApproval(c.Log, intUserID)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Render(groups)
}

// HandleGroup is the action method for approving/rejecting a group
func (c Requests) HandleGroup(groupID string, approve string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	// Authz check
	userModel, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Flash.Error("Unable to access user details")
		return c.Redirect(Account.Index)
	}

	if !userModel.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	intGroupID, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.Groups)
	}

	approveFlag, err := strconv.ParseBool(approve)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.Groups)
	}

	groupModel := &models.Group{
		GroupID: intGroupID,
	}

	// Push the group into database table
	err = groupModel.ApproveOrReject(c.Log, approveFlag)
	if err != nil {
		c.Flash.Error("Unable to add group at the moment. Please try after sometime.")
	}

	if approveFlag {
		c.Flash.Success("Group approved")
	} else {
		c.Flash.Success("Group rejected")
	}

	return c.Redirect(Requests.Groups)
}

// Users is the GET action for getting users pending approval
func (c Requests) Users() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	// Authorize
	user, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if user == nil {
		c.Log.Errorf("Unable to fetch user details from database")
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get all the groups applicable to user
	users, err := models.GetAllUsersPendingApproval(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch list of users. Error: %s", err.Error())
	}

	return c.Render(users)
}

// HandleUser is the action method for approving/rejecting a user
func (c Requests) HandleUser(memberID string, approve string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	// Authz check
	userModel, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Flash.Error("Unable to access user details")
		return c.Redirect(Account.Index)
	}

	if !userModel.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	intMemberID, err := strconv.ParseInt(memberID, 10, 64)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.Users)
	}

	approveFlag, err := strconv.ParseBool(approve)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.Users)
	}

	memberModel := &models.User{
		UserID: intMemberID,
	}

	// Push the group into database table
	err = memberModel.ApproveOrReject(c.Log, approveFlag)
	if err != nil {
		c.Flash.Error("Unable to process the request at the moment. Please try after sometime.")
	}

	if approveFlag {
		c.Flash.Success("User request approved")
	} else {
		c.Flash.Success("User request rejected")
	}

	return c.Redirect(Requests.Users)
}

// GroupAccess is the GET action for getting user-group mapping pending approval
func (c Requests) GroupAccess() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	// Authorize
	user, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if user == nil {
		c.Log.Errorf("Unable to fetch user details from database")
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get all the groups applicable to user
	userGroupMap, err := models.GetAllUserGroupMappingPendingApproval(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch list of users. Error: %s", err.Error())
	}

	return c.Render(userGroupMap)
}

// HandleGroupAccess is the action method for approving/rejecting a user
func (c Requests) HandleGroupAccess(groupID, memberID, approve string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUserName := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUserName

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	// Authz check
	userModel, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Flash.Error("Unable to access user details")
		return c.Redirect(Account.Index)
	}

	if !userModel.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	intMemberID, err := strconv.ParseInt(memberID, 10, 64)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.GroupAccess)
	}

	intGroupID, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.GroupAccess)
	}

	approveFlag, err := strconv.ParseBool(approve)
	if err != nil {
		c.Flash.Error("Internal error. Please try after sometime.")
		return c.Redirect(Requests.Users)
	}

	userGroupModel := &models.UserGroupMap{
		UserID:  intMemberID,
		GroupID: intGroupID,
	}

	// Push the group into database table
	err = userGroupModel.ApproveOrReject(c.Log, approveFlag)
	if err != nil {
		c.Flash.Error("Unable to process the request at the moment. Please try after sometime.")
	}

	if approveFlag {
		c.Flash.Success("Group access request approved")
	} else {
		c.Flash.Success("Group access request rejected")
	}

	return c.Redirect(Requests.GroupAccess)
}
