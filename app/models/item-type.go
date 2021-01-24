package models

import (
	"fmt"
	"strings"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/database"
)

var (
	supportedImageTypes = []string{"jpg", "jpeg", "png"}
	supportedVideoTypes = []string{"mp4"}
)

// ItemType is the struct for item types database table
type ItemType struct {
	tableName    struct{} `sql:"ItemTypes"`
	ItemTypeID   int      `sql:"item_type_id,pk"`
	ItemTypeName string   `sql:"item_type_name"`
	MaxItemCount int      `sql:"max_item_count"`
	MaxItemSpace float32  `sql:"max_item_space"`
}

// ItemLimits is the view model for item limits
type ItemLimits struct {
	MaxSizePictures float32
	MaxSizeVideos   float32
}

// GetItemTypeDetails returns the details of the item type as per the item-type ID
func GetItemTypeDetails(itemTypeID int) (*ItemType, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	itemType := &ItemType{
		ItemTypeID: itemTypeID,
	}
	err = client.GetPGClient().Select(itemType)
	if err != nil {
		return nil, err
	}

	return itemType, nil
}

// GetItemType returns the Item-type for a given file
func GetItemType(filename string) (common.ItemType, string, error) {
	parts := strings.Split(filename, ".")

	if len(parts) < 2 {
		return common.ItemTypeUnknown, "", fmt.Errorf("Invalid file type. File extension not available")
	}

	extension := parts[1]
	var itemType = common.ItemTypeUnknown
	for _, imageType := range supportedImageTypes {
		if imageType == extension {
			itemType = common.ItemTypePictures
			break
		}
	}

	for _, videoType := range supportedVideoTypes {
		if videoType == extension {
			itemType = common.ItemTypeVideos
			break
		}
	}

	if itemType == common.ItemTypeUnknown {
		return itemType, extension, fmt.Errorf("Invalid file type - '%s'. Supported types - pictures ('jpg', 'jpeg' and 'png') and videos ('mp4')", extension)
	}

	return itemType, extension, nil
}

// GetAllItemTypes returns the details of the item type as per the item-type ID
func GetAllItemTypes(log logger.MultiLogger) (*ItemLimits, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to get the item types")
	}

	var itemTypes []*ItemType
	err = client.GetPGClient().Model(&itemTypes).Select()
	if err != nil {
		log.Errorf("Unable to get all item types. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to get the item types")
	}

	itemLimits := &ItemLimits{}
	for _, item := range itemTypes {
		switch item.ItemTypeID {
		case common.ItemTypePictures.GetItemID():
			itemLimits.MaxSizePictures = item.MaxItemSpace
		case common.ItemTypeVideos.GetItemID():
			itemLimits.MaxSizeVideos = item.MaxItemSpace
		}
	}

	return itemLimits, nil
}

// UpdateLimits updates the item upload limits associated with a group
func (model *ItemType) UpdateLimits(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update group limits")
	}

	// Update item model in database
	res, err := client.GetPGClient().Model(model).
		Column("max_item_space").
		WherePK().
		Update()
	if err != nil {
		log.Errorf("Unable to update item limit into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to update item limits")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to update item limits")
	}

	return nil
}
