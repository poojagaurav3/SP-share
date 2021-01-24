package models

import (
	"fmt"
	"time"

	"github.com/go-pg/pg"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/database"
)

const (
	// MB is the const for conversion to MB from bytes (1 MB = 1e6 bytes)
	MB = 1e-06
)

// Item is the model for the metadata of an item added by a user
type Item struct {
	tableName    struct{}  `sql:"Items"`
	ItemID       int64     `sql:"item_id,pk"`
	ItemName     string    `sql:"item_name"`
	Description  string    `sql:"description"`
	ItemTypeID   int       `sql:"item_type_id"`
	ItemSize     int64     `sql:"item_size"`
	GroupID      int64     `sql:"group_id"`
	Uploaded     bool      `sql:"uploaded,default:false"`
	ItemPath     string    `sql:"item_path"`
	CreatedBy    int64     `sql:"created_by"`
	CreationTime time.Time `sql:"creation_time"`
	LastAccessed time.Time `sql:"last_accessed"`
}

// ItemView is the model for the metadata of an item to be used in the view
type ItemView struct {
	tableName          struct{}  `sql:"Items,alias:item"`
	ItemID             int64     `sql:"item_id,pk"`
	ItemName           string    `sql:"item_name"`
	Description        string    `sql:"description"`
	ItemTypeID         int       `sql:"item_type_id"`
	ItemSize           int64     `sql:"item_size"`
	GroupID            int64     `sql:"group_id"`
	Uploaded           bool      `sql:"uploaded,default:false"`
	ItemPath           string    `sql:"item_path"`
	CreatedByFirstName string    `sql:"created_by_first_name"`
	CreatedByLastName  string    `sql:"created_by_last_name"`
	CreatedBy          int64     `sql:"created_by"`
	CreationTime       time.Time `sql:"creation_time"`
	LastAccessed       time.Time `sql:"last_accessed"`
}

// ItemWithComments holds the item details with all comments
type ItemWithComments struct {
	ItemMeta  *ItemView
	GroupName string
	Comments  []*CommentDisplay
}

// Add adds the metadata of an item to database
func (model *Item) Add(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	// Insert the item model into database
	res, err := client.GetPGClient().Model(model).Returning("*").OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Errorf("Unable to insert item into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to add the item at the moment")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to add the item at the moment")
	}

	return nil
}

// MarkItemAsUploaded updates the upload status of the item to true
func MarkItemAsUploaded(log logger.MultiLogger, itemID int64) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	model := &Item{
		ItemID: itemID,
	}

	res, err := client.GetPGClient().Model(model).WherePK().Set("uploaded = ?", true).Update()
	if err != nil {
		log.Errorf("Unable to update the upload status of the item (ID: %d). Err: %s", itemID, err.Error())
		return fmt.Errorf("Unable to process the request")
	}
	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to process the request")
	}

	return nil
}

// GetItemsByGroupIDs returns all the item objects corresponding to the group ids provided
func GetItemsByGroupIDs(groupIDs []int64) ([]*Item, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	// Get all items for the group IDs provided
	var items []*Item
	err = client.GetPGClient().Model(&items).
		Where("group_id in (?)", pg.Ints(groupIDs)).
		Where("uploaded = ?", true).
		Select()
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetItemDetailsWithItemID returns the metadata of the item along with comments
func GetItemDetailsWithItemID(log logger.MultiLogger, itemID int64) (*ItemWithComments, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	itemMeta := &ItemView{
		ItemID: itemID,
	}

	// Get Item metadata
	err = client.GetPGClient().Model(itemMeta).WherePK().
		ColumnExpr(`"item".*`).
		ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
		Join("JOIN appuser AS u").
		JoinOn("u.user_id = \"item\".created_by").
		Select()
	if err != nil {
		log.Errorf("Unable to get item metadata. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	// Get all the comments
	comments, err := GetCommentsForAnItem(log, itemID)
	if err != nil {
		// error is already logged
		return nil, err
	}

	itemWithComments := &ItemWithComments{
		ItemMeta: itemMeta,
		Comments: comments,
	}

	return itemWithComments, nil
}

// CheckLimits checks whether the uploaded file satisfies the per user, per group and per item limits
func (model *Item) CheckLimits(log logger.MultiLogger) error {
	if model == nil {
		return fmt.Errorf("Item details unavailable")
	}

	// Get the filesize
	fileSizeInMB := float32(model.ItemSize) * MB

	// Check per-user limit
	err := model.checkPerUserLimits(log, fileSizeInMB)
	if err != nil {
		return err
	}

	// Check per-group limit
	err = model.checkPerGroupLimit(log, fileSizeInMB)
	if err != nil {
		return err
	}

	// Check per-item limit (Also validates the item type)
	err = model.checkPerItemTypeLimit(log, fileSizeInMB)
	if err != nil {
		return err
	}

	return nil
}

func (model *Item) checkPerUserLimits(log logger.MultiLogger, fileSizeInMB float32) error {
	// Get the limits tagged to the user
	user, err := GetUserByUserID(model.CreatedBy)
	if err != nil {
		log.Errorf("Unable to get upload limits for the user. Error: %s", err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the user")
	}

	// Get the aggregated count of items uploaded and space utilized by the user
	count, size, err := getLimitUtilizationForUser(model.CreatedBy)
	log.Infof("[User Limits] Count = %d, Size = %f, Uploaded file size = %f MB", count, size, fileSizeInMB)
	if err != nil {
		log.Errorf(err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the user")
	}

	if !(count < user.MaxItemCount) {
		return fmt.Errorf("User is limited to upload only %d items", user.MaxItemCount)
	}

	if !(size+fileSizeInMB < user.MaxItemSpace) {
		return fmt.Errorf("User is limited to %.3f MB of space for uploads", user.MaxItemSpace)
	}

	return nil
}

func (model *Item) checkPerGroupLimit(log logger.MultiLogger, fileSizeInMB float32) error {
	// Get the limits tagged to the user
	group, err := GetGroupDetailUsingID(model.GroupID)
	if err != nil {
		log.Errorf("Unable to get upload limits for the user. Error: %s", err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the user")
	}

	// Get the aggregated count of items uploaded and space utilized by the user
	count, size, err := getLimitUtilizationForGroup(model.GroupID)
	log.Infof("[Group Limits] Count = %d, Size = %f, Uploaded file size = %f MB", count, size, fileSizeInMB)
	if err != nil {
		log.Errorf(err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the user")
	}

	if !(count < group.MaxItemCount) {
		return fmt.Errorf("Only %d items can be uploaded in the group", group.MaxItemCount)
	}

	if !(size+fileSizeInMB < group.MaxItemSpace) {
		return fmt.Errorf("The group is limited to %.3f MB of space for uploads", group.MaxItemSpace)
	}

	return nil
}

func (model *Item) checkPerItemTypeLimit(log logger.MultiLogger, fileSizeInMB float32) error {
	// Get the limits tagged to the user
	itemType, err := GetItemTypeDetails(model.ItemTypeID)
	if err != nil {
		log.Errorf("Unable to get upload limits for the item type. Error: %s", err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the item type")
	}

	// Get the aggregated count of items uploaded and space utilized by the user
	_, size, err := getLimitUtilizationForGroup(model.GroupID)
	if err != nil {
		log.Errorf(err.Error())
		return fmt.Errorf("Unable to fetch upload limits for the item type")
	}
	log.Infof("[Item Type Limits] Size = %f, Uploaded file size = %f MB", size, fileSizeInMB)

	if fileSizeInMB > itemType.MaxItemSpace {
		return fmt.Errorf("Maximum allowed file size for item-type - '%s' is %.3f MB", itemType.ItemTypeName, itemType.MaxItemSpace)
	}

	return nil
}

func getLimitUtilizationForUser(userID int64) (int, float32, error) {
	var count int
	var totalSize float32

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	err = client.GetPGClient().Model((*Item)(nil)).
		ColumnExpr("count(*) AS count, sum(item_size) AS totalSize").
		Where("created_by = ?", userID).
		Select(&count, &totalSize)

	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get the utilized limits for the user - %d. Error: %s", userID, err.Error())
	}

	return count, totalSize * MB, nil
}

func getLimitUtilizationForGroup(groupID int64) (int, float32, error) {
	var count int
	var totalSize float32

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	err = client.GetPGClient().Model((*Item)(nil)).
		ColumnExpr("count(*) AS count, sum(item_size) AS totalSize").
		Where("group_id = ?", groupID).
		Select(&count, &totalSize)

	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get the utilized limits for the group - %d. Error: %s", groupID, err.Error())
	}

	return count, totalSize * MB, nil
}

func getLimitUtilizationForItemType(itemTypeID int) (int, float32, error) {
	var count int
	var totalSize float32

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	err = client.GetPGClient().Model((*Item)(nil)).
		ColumnExpr("count(*) AS count, sum(item_size) AS totalSize").
		Where("item_type_id = ?", itemTypeID).
		Select(&count, &totalSize)

	if err != nil {
		return -1, -1, fmt.Errorf("Unable to get the utilized limits for the item type - %d. Error: %s", itemTypeID, err.Error())
	}

	// Size is already in MB for per item limits
	return count, totalSize, nil
}

// GetItemDetailsByID returns the metadata of the item
func GetItemDetailsByID(log logger.MultiLogger, itemID int64) (*Item, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	itemMeta := &Item{
		ItemID: itemID,
	}

	// Get Item metadata
	err = client.GetPGClient().Select(itemMeta)
	if err != nil {
		log.Errorf("Unable to get item metadata. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	return itemMeta, nil
}

// Delete deletes the metadata of an item from database
func (model *Item) Delete(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	// Insert the item model into database
	res, err := client.GetPGClient().Model(model).WherePK().Delete()
	if err != nil {
		log.Errorf("Unable to delete item from the database. Err: %s", err.Error())
		return fmt.Errorf("Unable to delete the item at the moment")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to delete the item at the moment")
	}

	return nil
}

// MarkItemAsDeleting updates the upload status of the item to true
func MarkItemAsDeleting(log logger.MultiLogger, itemID int64) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	model := &Item{
		ItemID: itemID,
	}

	res, err := client.GetPGClient().Model(model).WherePK().Set("uploaded = ?", false).Update()
	if err != nil {
		log.Errorf("Unable to update the upload status of the item (ID: %d). Err: %s", itemID, err.Error())
		return fmt.Errorf("Unable to process the request")
	}
	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to process the request")
	}

	return nil
}
