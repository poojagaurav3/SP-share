package auth

import (
	"strconv"

	"github.com/revel/revel"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/controllers"
	"github.com/sp-share/app/models"
)

// Authenticate is the function interceptor for authentication checks
func Authenticate(c *revel.Controller) revel.Result {
	user := &models.User{}
	_, err := c.Session.GetInto("user", user, false)
	if err != nil {
		c.Log.Error("Unauthenticated. Unable to fetch user from session. Err: %+v", err)
		return c.Redirect(controllers.Account.Index)
	}

	c.Flash.Out["userID"] = strconv.FormatInt(user.GetUserID(), 10)
	c.Flash.Out["loggedInUser"] = user.GetUserName()
	if user.WorkflowStatus == common.WorkflowStatusPending.GetStatusID() {
		return c.Redirect(controllers.Account.Unauthorized)
	}
	return nil
}
