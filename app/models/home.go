package models

import (
	"fmt"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/common"
)

// HomeView is the model for the home page
type HomeView struct {
	Images []*ItemInGroup
	Videos []*ItemInGroup
}

// ItemInGroup contains item metadata along with group tagging
type ItemInGroup struct {
	GroupDetails *GroupKeyVal
	ItemMeta     *Item
}

// GetHomePageData get the data for home page for the logged in user
func GetHomePageData(log logger.MultiLogger, userID int64) (*HomeView, error) {
	homeViewModel := &HomeView{}

	// Get all the groups for the logged in user
	groups, err := GetAllGroupsKeyVal(userID)
	if err != nil {
		// We already logged this error
		return nil, err
	}
	if len(groups) == 0 {
		// There are no groups present at this point
		return nil, nil
	}

	groupKeyValMap := make(map[int64]*GroupKeyVal)

	// Fetch all the group IDs
	groupIDs := make([]int64, len(groups))
	for index, group := range groups {
		groupIDs[index] = group.GroupID

		// Add to mapping for id->name for group
		groupKeyValMap[group.GroupID] = group
	}

	// Get all items
	items, err := GetItemsByGroupIDs(groupIDs)
	if err != nil {
		log.Errorf("Unable to get item metadata from database. Error: %+v", err)
		return nil, fmt.Errorf("Unable to fetch the item metadata")
	}

	for _, item := range items {
		// Get the group object
		groupObj, present := groupKeyValMap[item.GroupID]
		if !present {
			// This should not happen
			// If it does, we'll skip this item and log the error
			log.Errorf("Didn't find the group object in the group map [GroupID: %d]", item.GroupID)
			continue
		}

		itemInGroup := &ItemInGroup{
			GroupDetails: groupObj,
			ItemMeta:     item,
		}

		switch item.ItemTypeID {
		case common.ItemTypePictures.GetItemID():
			homeViewModel.Images = append(
				homeViewModel.Images,
				itemInGroup,
			)
		case common.ItemTypeVideos.GetItemID():
			homeViewModel.Videos = append(
				homeViewModel.Videos,
				itemInGroup,
			)
		}
	}

	return homeViewModel, nil
}
