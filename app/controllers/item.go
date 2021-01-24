package controllers

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/revel/revel"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/models"
)

// Item is the controller for item uploads/downloads
type Item struct {
	*revel.Controller
}

// Upload is the GET action for upload form for items (photos and videos)
func (c Item) Upload() revel.Result {
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
	groupsKeyVal, err := models.GetAllGroupsKeyVal(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to fetch list of groups. Error: %s", err.Error())
		c.Flash.Error("Could not load data")
	}

	return c.Render(groupsKeyVal)
}

// UploadHandler takes care of the file uploads
// Supported files: Images (.jpeg, .jpg, .png) and videos (TODO)
func (c Item) UploadHandler(uploadedFile []byte, group, description, name string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
	}

	c.Validation.Required(uploadedFile).Message("File is required")
	c.Validation.Required(group).Message("Group name is required")
	c.Validation.Required(name).Message("Item name is required")
	c.Validation.Required(description).Message("Description is required")
	c.Validation.MaxSize(name, 30).Message("Item name should be less than 30 characters")
	c.Validation.MaxSize(description, 400).Message("Description should be less than 400 characters")
	c.Validation.Match(name, regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]+$")).Message("Item name should start with an alphabet and must include only alphabets (a-z, A-Z), numbers (0-9) and symbols (_)")

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Item.Upload)
	}

	// In case of validation errors, pass them on to the UI
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Item.Upload)
	}

	intGroupID, err := strconv.ParseInt(group, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid Group ID found - %s. Error: %s", group, err.Error())
		c.Flash.Error("Unable to process the request")
		return c.Redirect(Group.Details)
	}

	file := c.Params.Files["uploadedFile"][0]

	// Verify Item-type
	itemType, _, err := models.GetItemType(file.Filename)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Item.Upload)
	}

	fileNameHashed := common.SHA256(fmt.Sprintf("%s%d", file.Filename, time.Now().Unix()))
	c.Log.Infof("filename hash: %s", fileNameHashed)
	filePath := fmt.Sprintf("./uploads/%s", fileNameHashed)

	fileReader, err := file.Open()
	if err != nil {
		c.Flash.Error("Invalid file")
		return c.Redirect(Item.Upload)
	}
	defer fileReader.Close()

	itemModel := &models.Item{
		ItemName:    name,
		Description: description,
		ItemPath:    filePath[1:],
		ItemSize:    file.Size,
		ItemTypeID:  itemType.GetItemID(),
		GroupID:     intGroupID,
		CreatedBy:   intUserID,
	}

	// Check limits
	err = itemModel.CheckLimits(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Item.Upload)
	}

	// Add the item to database with status as 'uploaded=false'
	err = itemModel.Add(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Item.Upload)
	}

	// Upload the item to S3
	fileBytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		c.Flash.Error("Cannot read file")
		return c.Redirect(Item.Upload)
	}
	// Create a new file
	err = ioutil.WriteFile(filePath, fileBytes, 0644)
	if err != nil {
		c.Flash.Error("Unable to write to temp file. Error: %+v", err)
		return c.Redirect(Item.Upload)
	}

	// Upload the status of the item in the database to 'uploaded=true'
	err = models.MarkItemAsUploaded(c.Log, itemModel.ItemID)
	if err != nil {
		c.Flash.Error(err.Error())
	} else {
		c.Flash.Success("Successfully uploaded the file")
	}

	return c.Redirect(Item.Upload)
}

// Preview is the GET action for item details
func (c Item) Preview(id int) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	// Get all the groups for Authz check
	groups, err := models.GetAllGroupsKeyVal(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to get the groups for user - %d. Error: %s", intUserID, err.Error())
		c.Flash.Error("Unable to get the details of the item")
		return c.Render(nil)
	}

	// Get Item with Comments
	itemWithComments, err := models.GetItemDetailsWithItemID(c.Log, int64(id))
	if err != nil {
		c.Flash.Error(err.Error())
	}

	// Check if user has access to the item
	if itemWithComments == nil ||
		itemWithComments.ItemMeta == nil {
		c.Flash.Error("Item details unavailable")
		return c.Redirect(Home.Index)
	}

	exists, groupName := checkIfGroupIDExists(groups, itemWithComments.ItemMeta.GroupID)
	if !exists {
		c.Flash.Error("Unauthorized! You do not have enough permissions to view the content")
		return c.Redirect(Home.Index)
	}

	itemWithComments.GroupName = groupName

	return c.Render(itemWithComments)
}

// AddComment adds a comment to the given item
func (c Item) AddComment(itemID string, comment string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	intItemID, err := strconv.ParseInt(itemID, 10, 64)
	if err != nil {
		c.Log.Errorf("Unable to parse the item ID - %s. Error: %s", itemID, err.Error())
		c.Flash.Error("Unable to add the comment")
		return c.Redirect(Home.Index)
	}

	// Escape invalid characters
	escapedComment := html.EscapeString(comment)

	commentObj := &models.Comment{
		Comment:   escapedComment,
		ItemID:    intItemID,
		CreatedBy: intUserID,
	}

	err = commentObj.Add(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
	}

	return c.Redirect("/item/%d", intItemID)
}

// Delete adds a comment to the given item
func (c Item) Delete(itemID string) revel.Result {
	userID := c.Flash.Out["userID"]
	loggedInUser := c.Flash.Out["loggedInUser"]
	c.Flash.Out["loggedInUser"] = loggedInUser

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.Log.Errorf("Invalid User ID found in session - %s. Error: %s", userID, err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	intItemID, err := strconv.ParseInt(itemID, 10, 64)
	if err != nil {
		c.Log.Errorf("Unable to parse the item ID - %s. Error: %s", itemID, err.Error())
		c.Flash.Error("Unable to add the comment")
		return c.Redirect(Home.Index)
	}

	user, err := models.GetUserByUserID(intUserID)
	if err != nil {
		c.Log.Errorf("Unable to get user details. Error: %s", err.Error())
		c.Flash.Error("Please login to continue")
		return c.Redirect(Account.Index)
	}

	// Get Item details
	itemMeta, err := models.GetItemDetailsByID(c.Log, intItemID)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Home.Index)
	}

	if itemMeta.CreatedBy != intUserID && !user.IsAdmin {
		c.Flash.Error("Unauthorized. You do not have enough permissions to delete the item.")
		return c.Redirect(Home.Index)
	}

	err = models.MarkItemAsDeleting(c.Log, itemMeta.ItemID)
	if err != nil {
		c.Flash.Error("Unable to delete the item at the moment")
		return c.Redirect(Home.Index)
	}

	// Delete file
	err = os.Remove(fmt.Sprintf("./%s", itemMeta.ItemPath))
	if err != nil {
		c.Flash.Error("Unable to delete the item at the moment")
		return c.Redirect(Home.Index)
	}

	// Delete the item
	err = itemMeta.Delete(c.Log)
	if err != nil {
		c.Flash.Error(err.Error())
		return c.Redirect(Home.Index)
	}

	c.Flash.Success("Successfully deleted the item")
	return c.Redirect(Home.Index)
}
