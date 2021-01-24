package controllers

import (
	"strconv"

	"github.com/revel/revel"
	"github.com/sp-share/app/models"
)

// Home is the controller home page functions
type Home struct {
	*revel.Controller
}

// Index is the GET action for Home/Index page
func (c Home) Index() revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	homeItems, err := models.GetHomePageData(c.Log, intUserID)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Render(homeItems)
}

// checkIfGroupIDExists checks whether a groupID exists in list of groups
func checkIfGroupIDExists(allGroups []*models.GroupKeyVal, groupID int64) (bool, string) {
	for _, group := range allGroups {
		if group.GroupID == groupID {
			return true, group.GroupName
		}
	}

	return false, ""
}
