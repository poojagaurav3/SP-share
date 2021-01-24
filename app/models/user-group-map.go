package models

import (
	"fmt"
	"time"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/database"
)

// UserGroupMap holds the mapping of users and groups
type UserGroupMap struct {
	tableName      struct{}  `sql:"UserGroupMap"`
	UserID         int64     `sql:"user_id,pk"`
	GroupID        int64     `sql:"group_id,pk"`
	IsLeader       bool      `sql:"is_leader,default:false"`
	CreatedBy      int64     `sql:"created_by"`
	CreationTime   time.Time `sql:"creation_time"`
	LastUpdated    time.Time `sql:"last_updated"`
	WorkflowStatus int       `sql:"workflow_status"`
}

// UserGroupMapView is the display model for UserGroupMap
type UserGroupMapView struct {
	tableName          struct{}  `sql:"UserGroupMap,alias:ugm"`
	UserID             int64     `sql:"user_id"`
	GroupID            int64     `sql:"group_id"`
	GroupName          string    `sql:"group_name"`
	IsLeader           bool      `sql:"is_leader"`
	Username           string    `sql:"username"`
	FirstName          string    `sql:"first_name"`
	LastName           string    `sql:"last_name"`
	CreatedByFirstName string    `sql:"created_by_first_name"`
	CreatedByLastName  string    `sql:"created_by_last_name"`
	CreationTime       time.Time `sql:"creation_time"`
	WorkflowStatus     int       `sql:"workflow_status"`
}

// Add adds a user-group mapping into database
func (model *UserGroupMap) Add(log logger.MultiLogger, userID int64) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	// Check if user has sufficient previleges
	existingUserGroupMap := &UserGroupMap{
		UserID:  userID,
		GroupID: model.GroupID,
	}

	err = client.GetPGClient().Select(existingUserGroupMap)
	if err != nil {
		log.Errorf("Unable to fetch user permissions. Error: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}

	if !existingUserGroupMap.IsLeader {
		return fmt.Errorf("You do not have sufficient privileges to add users to the group")
	}

	// Insert group model into database
	model.WorkflowStatus = common.WorkflowStatusApproved.GetStatusID()
	res, err := client.GetPGClient().Model(model).OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Errorf("Unable to insert user into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}
	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to process the request")
	}

	return nil
}

// UpgradeToGroupLead upgrades a given mapping to include group lead access
func (model *UserGroupMap) UpgradeToGroupLead(log logger.MultiLogger) error {
	if model == nil {
		log.Errorf("UserGroupMap model is nil")
		return fmt.Errorf("Unable to perform the action at the moment")
	}
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	// Check if the mapping exits
	newModel := &UserGroupMap{
		UserID:  model.UserID,
		GroupID: model.GroupID,
	}

	err = client.GetPGClient().Select(newModel)
	if err != nil {
		log.Errorf("Unable to update the user-group mapping. Error: %s", err.Error())
		return fmt.Errorf("User does not have access to the group")
	}

	wfStatus := common.WorkflowStatusPending.GetStatusID()
	res, err := client.GetPGClient().Model(model).WherePK().
		Set("workflow_status = ?", wfStatus).
		Set("is_leader = ?", true).
		Update()
	if err != nil {
		log.Errorf("Unable to update the user-group mapping. Error: %s", err.Error())
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	return nil
}

// GetAllUserGroupMappingForAGroup returns list of all the users that are pending for admin approval
func GetAllUserGroupMappingForAGroup(groupID int64) ([]*UserGroupMapView, error) {
	var users []*UserGroupMapView

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	// Get all the groups
	err = client.GetPGClient().Model(&users).
		ColumnExpr(`"ugm".user_id, "ugm".group_id, "ugm".is_leader, "ugm".creation_time, "ugm".workflow_status`).
		ColumnExpr(`u.first_name, u.last_name, u.username`).
		ColumnExpr(`au.first_name AS created_by_first_name, au.last_name AS created_by_last_name`).
		Join("JOIN appuser AS u").
		JoinOn("u.user_id = \"ugm\".user_id").
		Join("JOIN appuser AS au").
		JoinOn("au.user_id = \"ugm\".created_by").
		Where("\"ugm\".group_id = ?", groupID).Select()

	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetAllUserGroupMappingPendingApproval returns list of all the user-group mapping requests that are pending for admin approval
func GetAllUserGroupMappingPendingApproval(userID int64) ([]*UserGroupMapView, error) {
	var users []*UserGroupMapView

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	// Get the User
	userModel, err := GetUserByUserID(userID)
	if err != nil {
		return nil, err
	}

	if !userModel.IsAdmin {
		return nil, fmt.Errorf("unauthorized - only admins can access the content")
	}

	actionID := common.WorkflowStatusPending.GetStatusID()
	// Get all the user-group mappings with pending status
	err = client.GetPGClient().Model(&users).
		ColumnExpr(`"ugm".user_id, "ugm".group_id, "ugm".is_leader, "ugm".creation_time, "ugm".workflow_status`).
		ColumnExpr(`u.first_name, u.last_name, u.username`).
		ColumnExpr(`au.first_name AS created_by_first_name, au.last_name AS created_by_last_name`).
		ColumnExpr(`g.group_name`).
		Join("JOIN appuser AS u").
		JoinOn("u.user_id = \"ugm\".user_id").
		Join("JOIN appuser AS au").
		JoinOn("au.user_id = \"ugm\".created_by").
		Join("JOIN groups AS g").
		JoinOn("g.group_id = \"ugm\".group_id and g.created_by <> \"ugm\".user_id").
		Where("\"ugm\".workflow_status = ?", actionID).Select()

	if err != nil {
		return nil, err
	}

	return users, nil
}

// ApproveOrReject is used by admin to approve/reject a group creation
func (model *UserGroupMap) ApproveOrReject(log logger.MultiLogger, approve bool) error {
	// Note: Authz check is already done at this point

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update the group at the moment")
	}

	// Get the group object
	groupModel := &Group{
		GroupID: model.GroupID,
	}

	err = client.GetPGClient().Select(groupModel)
	if err != nil {
		log.Errorf("Unable to fetch the group details. Error: %s", err.Error())
		return fmt.Errorf("Unable to fetch the group details at the moment")
	}

	existingMapping := &UserGroupMap{
		UserID:  model.UserID,
		GroupID: model.GroupID,
	}
	err = client.GetPGClient().Select(existingMapping)
	if err != nil {
		log.Errorf("Unable to update the group details. Error: %s", err.Error())
		return fmt.Errorf("The given user does not have access to group '%s'", groupModel.GroupName)
	}

	if approve {
		model.WorkflowStatus = common.WorkflowStatusApproved.GetStatusID()

		res, err := client.GetPGClient().Model(model).WherePK().Set("workflow_status = ?", model.WorkflowStatus).Update()
		if err != nil {
			log.Errorf("Unable to update the user-group mapping. Error: %s", err.Error())
			return fmt.Errorf("Unable to perform the action at the moment")
		}

		if res.RowsAffected() < 1 {
			return fmt.Errorf("Unable to perform the action at the moment")
		}
	} else {
		// Check the request type - Request for a new group, or request for access upgrade
		if groupModel.CreatedBy == model.UserID {
			// Group was created by the requesting user
			// This request is handled through the 'Create Group' request
			return fmt.Errorf("This request is associated with new group. Please check the 'Create Group' requests")
		}

		if existingMapping.IsLeader {
			// This is a request for upgrade
			// We downgrade the request to 'member' in this case
			model.WorkflowStatus = common.WorkflowStatusApproved.GetStatusID()
			res, err := client.GetPGClient().Model(model).WherePK().
				Set("workflow_status = ?", model.WorkflowStatus).
				Set("is_leader = ?", false).
				Update()
			if err != nil {
				log.Errorf("Unable to update the user-group mapping. Error: %s", err.Error())
				return fmt.Errorf("Unable to perform the action at the moment")
			}

			if res.RowsAffected() < 1 {
				return fmt.Errorf("Unable to perform the action at the moment")
			}
		}
	}

	return nil
}
