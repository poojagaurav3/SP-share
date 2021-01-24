package controllers

import (
	"strconv"

	"github.com/revel/revel"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/models"
)

// Limit is the controller for configuring item limits
type Limit struct {
	*revel.Controller
}

// Users sets the user level limit
func (c Limit) Users() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

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
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get list of all users
	users, err := models.GetAllUsersKeyVal(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Render(users)
}

// UserLimits gets the user limits associated with a given user
func (c Limit) UserLimits(user string) revel.Result {
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
	loggedInUser, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if loggedInUser == nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !loggedInUser.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	c.Validation.Required(user).Message("User is required")
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Limit.Users)
	}

	intUser, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intUser, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	userObj, err := models.GetUserByUserID(intUser)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Limit.Users)
	}

	// Get list of all users
	users, err := models.GetAllUsersKeyVal(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	userLimitsModel := &models.UserLimits{
		UserID:       intUser,
		Users:        users,
		MaxItemCount: userObj.MaxItemCount,
		MaxItemSpace: userObj.MaxItemSpace,
	}

	return c.Render(userLimitsModel)
}

// UpdateUserLimits updates the user limit in database
func (c Limit) UpdateUserLimits(user, maxCount, maxSpace string) revel.Result {
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
	loggedInUser, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if loggedInUser == nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !loggedInUser.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	intUser, err := strconv.ParseInt(user, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intUser, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	intMaxCount, err := strconv.ParseInt(maxCount, 10, 32)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intUser, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	intMaxSpace, err := strconv.ParseFloat(maxSpace, 32)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intUser, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	userToUpdate := &models.User{
		UserID:       intUser,
		MaxItemCount: int(intMaxCount),
		MaxItemSpace: float32(intMaxSpace),
	}

	err = userToUpdate.UpdateLimits(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Limit.Users)
	}

	c.Flash.Success("Updated user limits")
	return c.Redirect(Limit.Users)
}

//Groups sets the group level limit
func (c Limit) Groups() revel.Result {
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
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get list of all groups
	groups, err := models.GetAllPresentGroupsKeyVal(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Render(groups)
}

// GroupLimits gets the group limits
func (c Limit) GroupLimits(group string, user string) revel.Result {
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
	loggedInUser, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if loggedInUser == nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !loggedInUser.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	c.Validation.Required(group).Message("Group is required")
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Limit.Groups)
	}

	intGroup, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid Group ID found - %s. Error: %s", intGroup, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	groupObj, err := models.GetGroupDetailUsingID(intGroup)
	if err != nil {
		c.Log.Errorf("Unable to fetch Group details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Limit.Groups)
	}

	// Get list of all groups
	groups, err := models.GetAllPresentGroupsKeyVal(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	groupLimitsModel := &models.GroupLimits{
		GroupID:      intGroup,
		Groups:       groups,
		MaxItemCount: groupObj.MaxItemCount,
		MaxItemSpace: groupObj.MaxItemSpace,
	}

	return c.Render(groupLimitsModel)
}

// UpdateGroupLimits updates the group limit in database
func (c Limit) UpdateGroupLimits(group, maxCount, maxSpace string) revel.Result {
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
	loggedInUser, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if loggedInUser == nil {
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !loggedInUser.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	intGroup, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid Group ID found - %s. Error: %s", intGroup, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	intMaxCount, err := strconv.ParseInt(maxCount, 10, 32)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intGroup, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	intMaxSpace, err := strconv.ParseFloat(maxSpace, 32)
	if err != nil {
		c.Log.Errorf("Invalid User ID found - %s. Error: %s", intGroup, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	groupToUpdate := &models.Group{
		GroupID:      intGroup,
		MaxItemCount: int(intMaxCount),
		MaxItemSpace: float32(intMaxSpace),
	}

	err = groupToUpdate.UpdateLimits(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Limit.Groups)
	}

	c.Flash.Success("Updated group limits")
	return c.Redirect(Limit.Groups)
}

// ItemTypes sets the per item-type limit
func (c Limit) ItemTypes() revel.Result {
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
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	// Get list of all groups
	itemLimits, err := models.GetAllItemTypes(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Render(itemLimits)
}

// UpdateItemLimits updates the per item-type limit
func (c Limit) UpdateItemLimits(maxSizePicture, maxSizeVideo string) revel.Result {
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
		c.Log.Errorf("Unable to fetch user details from database. Error: %s", err.Error())
		c.Flash.Error("Unable to fetch user details")
		return c.Redirect(Home.Index)
	}

	if !user.IsAdmin {
		return c.Redirect(Account.Unauthorized)
	}

	flMaxSpacePictures, err := strconv.ParseFloat(maxSizePicture, 32)
	if err != nil {
		c.Log.Errorf("Invalid size for item type picture. Error: %s", err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	flMaxSpaceVideos, err := strconv.ParseFloat(maxSizeVideo, 32)
	if err != nil {
		c.Log.Errorf("Invalid size for item type videos. Error: %s", err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Home.Index)
	}

	itemTypeImage := &models.ItemType{
		ItemTypeID:   common.ItemTypePictures.GetItemID(),
		MaxItemSpace: float32(flMaxSpacePictures),
	}

	err = itemTypeImage.UpdateLimits(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Limit.ItemTypes)
	}

	itemTypeVideo := &models.ItemType{
		ItemTypeID:   common.ItemTypeVideos.GetItemID(),
		MaxItemSpace: float32(flMaxSpaceVideos),
	}

	err = itemTypeVideo.UpdateLimits(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Limit.ItemTypes)
	}

	c.Flash.Success("Limits updated successfully")
	return c.Redirect(Limit.ItemTypes)
}
