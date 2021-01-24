package models

import (
	"fmt"
	"time"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/database"
)

// Comment is the model for comments database table
type Comment struct {
	tableName    struct{}  `sql:"Comments,alias:c"`
	CommentID    int64     `sql:"comment_id,pk"`
	Comment      string    `sql:"comment"`
	ItemID       int64     `sql:"item_id"`
	CreatedBy    int64     `sql:"created_by"`
	CreationTime time.Time `sql:"creation_time"`
}

// CommentDisplay is the model for comments for frontend
type CommentDisplay struct {
	tableName          struct{}  `sql:"Comments,alias:c"`
	CommentID          int64     `sql:"comment_id,pk"`
	Comment            string    `sql:"comment"`
	ItemID             int64     `sql:"item_id"`
	CreatedBy          int64     `sql:"created_by"`
	CreatedByFirstName string    `sql:"created_by_first_name"`
	CreatedByLastName  string    `sql:"created_by_last_name"`
	CreationTime       time.Time `sql:"creation_time"`
}

// GetCommentsForAnItem returns all the comments for an item
func GetCommentsForAnItem(log logger.MultiLogger, itemID int64) ([]*CommentDisplay, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	var comments []*CommentDisplay

	query := client.GetPGClient().Model(&comments)
	query = query.ColumnExpr(`"c".comment, "c".creation_time`).
		ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
		Join("JOIN appuser AS u").
		JoinOn("u.user_id = \"c\".created_by").
		Where("item_id = ?", itemID).
		Order("creation_time ASC")

	err = query.Select()
	if err != nil {
		log.Errorf("Unable to get comments for the item - %d. Err: %s", itemID, err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	return comments, nil
}

// Add adds the comment to the database
func (model *Comment) Add(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get the database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	res, err := client.GetPGClient().Model(model).OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Errorf("Unable to insert the comment into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to process the request")
	}

	return nil
}
