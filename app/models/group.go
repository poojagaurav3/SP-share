package models

import (
	"fmt"
	"time"

	"github.com/revel/revel/logger"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/database"
)

// Group is the model for user groups present in SP Share
type Group struct {
	tableName      struct{}  `sql:"Groups"`
	GroupID        int64     `sql:"group_id,pk"`
	GroupName      string    `sql:"group_name"`
	CreatedBy      int64     `sql:"created_by"`
	CreationTime   time.Time `sql:"creation_time"`
	LastUpdated    time.Time `sql:"last_updated"`
	WorkflowStatus int       `sql:"workflow_status"`
	MaxItemCount   int       `sql:"max_item_count"`
	MaxItemSpace   float32   `sql:"max_item_space"`
}

// GroupKeyVal is the model for group key-val list (used for dropdowns etc)
type GroupKeyVal struct {
	tableName struct{} `sql:"Groups,alias:group"`
	GroupID   int64    `sql:"group_id,pk"`
	GroupName string   `sql:"group_name"`
	CreatedBy int64    `sql:"created_by"`
}

// GroupView is the group view model
type GroupView struct {
	tableName          struct{}  `sql:"Groups,alias:group"`
	GroupID            int64     `sql:"group_id,pk"`
	GroupName          string    `sql:"group_name"`
	CreatedByFirstName string    `sql:"created_by_first_name"`
	CreatedByLastName  string    `sql:"created_by_last_name"`
	CreationTime       time.Time `sql:"creation_time"`
	WorkflowStatus     int       `sql:"workflow_status"`
	IsLeader           bool      `sql:"is_leader"`
}

// GroupDetails is the group view model
type GroupDetails struct {
	tableName             struct{}  `sql:"Groups,alias:group"`
	GroupID               int64     `sql:"group_id,pk"`
	GroupName             string    `sql:"group_name"`
	CreatedByFirstName    string    `sql:"created_by_first_name"`
	CreatedByLastName     string    `sql:"created_by_last_name"`
	CreationTime          time.Time `sql:"creation_time"`
	WorkflowStatus        int       `sql:"workflow_status"`
	UserMapWorkflowStatus int       `sql:"user_map_workflow_status"`
	IsLeader              bool      `sql:"is_leader"`
	TaggedUsers           []*UserGroupMapView
}

// GroupLimits is the view model for setting group level limits
type GroupLimits struct {
	GroupID      int64
	Groups       []*GroupKeyVal
	MaxItemCount int
	MaxItemSpace float32
}

// Add inserts a group to database
func (model *Group) Add(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to insert group into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to get database client")
	}

	// TODO: Maybe try transactions

	// Insert group model into database
	_, err = client.GetPGClient().Model(model).Returning("*").OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Errorf("Unable to insert group into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to create user group at the moment")
	}

	// Also map the user to the newly created group
	userGroupMap := &UserGroupMap{
		UserID:         model.CreatedBy,
		GroupID:        model.GroupID,
		CreatedBy:      model.CreatedBy,
		IsLeader:       true,
		WorkflowStatus: common.WorkflowStatusApproved.GetStatusID(),
	}

	// Insert group model into database
	model.WorkflowStatus = common.WorkflowStatusPending.GetStatusID()
	res, err := client.GetPGClient().Model(userGroupMap).OnConflict("DO NOTHING").Insert()
	if err != nil {
		log.Errorf("Unable to insert user-mapping into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to process the request")
	}
	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to process the request")
	}

	return nil
}

// GetGroupDetails returns the details of the group using group ID and user ID
func GetGroupDetails(log logger.MultiLogger, userID int64, groupID int64) (*GroupDetails, error) {
	group := &GroupDetails{}

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	// Get the User
	userModel, err := GetUserByUserID(userID)
	if err != nil {
		log.Errorf("Unable to get user details. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	query := client.GetPGClient().Model(group)

	if !userModel.IsAdmin {
		// non-admin user
		query = query.ColumnExpr(`"group".group_id, "group".group_name, "group".workflow_status, "group".creation_time`).
			ColumnExpr(`ugm.is_leader, ugm.workflow_status AS user_map_workflow_status`).
			ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
			Join("JOIN usergroupmap AS ugm").
			JoinOn("ugm.group_id = \"group\".group_id").
			Join("JOIN appuser AS u").
			JoinOn("u.user_id = \"group\".created_by").
			Where("\"group\".group_id = ?", groupID).
			Where("ugm.user_id = ?", userID)
	} else {
		// admin
		query = query.ColumnExpr(`"group".group_id, "group".group_name, "group".workflow_status, "group".creation_time`).
			ColumnExpr(`true AS is_leader, 1 AS user_map_workflow_status`).
			ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
			Join("JOIN appuser AS u").
			JoinOn("u.user_id = \"group\".created_by").
			Where("\"group\".group_id = ?", groupID)
	}
	err = query.Select()
	if err != nil {
		log.Errorf("Unable to get group details. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to fetch the details of the Group")
	}
	if group.GroupID < 1 {
		log.Errorf("Unable to get group details")
		return nil, fmt.Errorf("The group does not exist")
	}

	if group.UserMapWorkflowStatus == common.WorkflowStatusPending.GetStatusID() {
		log.Info("Marking the user as not a leader")
		group.IsLeader = false
	}

	// Get all the users tagged to the group
	users, err := GetAllUserGroupMappingForAGroup(groupID)
	if err != nil {
		return group, fmt.Errorf("Unable to fetch the users tagged to the group")
	}
	group.TaggedUsers = users

	return group, err
}

// UpdateLimits updates the item upload limits associated with a group
func (model *Group) UpdateLimits(log logger.MultiLogger) error {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update group limits")
	}

	// Update user model in database
	res, err := client.GetPGClient().Model(model).
		Column("max_item_count").
		Column("max_item_space").
		WherePK().
		Update()
	if err != nil {
		log.Errorf("Unable to update group limit into database. Err: %s", err.Error())
		return fmt.Errorf("Unable to update group limits")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to update group limits")
	}

	return nil
}

// GetAllGroups returns list of all the groups that user has access to
// Approved as well as pending groups are returned
func GetAllGroups(userID int64) ([]*GroupView, error) {
	var groups []*GroupView

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

	query := client.GetPGClient().Model(&groups)

	if !userModel.IsAdmin {
		// Get only those groups in which user has access
		query = query.ColumnExpr(`"group".group_id, "group".group_name, "group".workflow_status, "group".creation_time`).
			ColumnExpr(`ugm.is_leader`).
			ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
			Join("JOIN usergroupmap AS ugm").
			JoinOn("ugm.group_id = \"group\".group_id").
			Join("JOIN appuser AS u").
			JoinOn("u.user_id = \"group\".created_by").
			Where("ugm.user_id = ?", userID)
	} else {
		query = query.ColumnExpr(`"group".group_id, "group".group_name, "group".workflow_status, "group".creation_time`).
			ColumnExpr(`true AS is_leader`).
			ColumnExpr(`u.first_name AS created_by_first_name, u.last_name AS created_by_last_name`).
			Join("JOIN appuser AS u").
			JoinOn("u.user_id = \"group\".created_by")
	}
	err = query.Select()

	return groups, err
}

// GetAllGroupsPendingApproval returns list of all the groups that are pending for admin approval
func GetAllGroupsPendingApproval(log logger.MultiLogger, userID int64) ([]*GroupView, error) {
	var groups []*Group

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	// Get the User
	userModel, err := GetUserByUserID(userID)
	if err != nil {
		log.Errorf("Unable to get user details. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to process the request")
	}

	if !userModel.IsAdmin {
		return nil, fmt.Errorf("unauthorized - only admins can access the content")
	}

	actionID := common.WorkflowStatusPending.GetStatusID()
	// Get all the groups
	client.GetPGClient().Model(&groups).Where("workflow_status = ?", actionID).Select()

	return convertGroupSliceToGroupViewSlice(groups), nil
}

func convertGroupSliceToGroupViewSlice(groups []*Group) []*GroupView {
	if groups == nil {
		return nil
	}

	result := make([]*GroupView, len(groups))
	for i, group := range groups {
		result[i] = convertGroupToGroupView(group)

		// wf := common.WorkflowStatus(group.WorkflowStatus)
		// result[i].WorkflowStatus = wf.GetString()
	}
	return result
}

func convertGroupToGroupView(group *Group) *GroupView {
	return &GroupView{
		GroupID:      group.GroupID,
		GroupName:    group.GroupName,
		CreationTime: group.CreationTime,
	}
}

// ApproveOrReject is used by admin to approve/reject a group creation
func (model *Group) ApproveOrReject(log logger.MultiLogger, approve bool) error {
	// Note: Authz check is already done at this point

	var actionID common.WorkflowStatus

	if approve {
		actionID = common.WorkflowStatusApproved
	} else {
		actionID = common.WorkflowStatusRejected
	}

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return fmt.Errorf("Unable to update the group at the moment")
	}

	model.WorkflowStatus = actionID.GetStatusID()
	res, err := client.GetPGClient().Model(model).WherePK().Set("workflow_status = ?", model.WorkflowStatus).Update()
	if err != nil {
		log.Errorf("Unable to update the group model. Error: %s", err.Error())
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	if res.RowsAffected() < 1 {
		return fmt.Errorf("Unable to perform the action at the moment")
	}

	return nil
}

// GetAllGroupsKeyVal returns list of all the groups that user has access to
// Only approved groups are returned
func GetAllGroupsKeyVal(userID int64) ([]*GroupKeyVal, error) {
	var groups []*GroupKeyVal

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

	query := client.GetPGClient().Model(&groups)
	if !userModel.IsAdmin {
		query = query.ColumnExpr(`"group".group_id, "group".group_name`).
			Join("JOIN usergroupmap AS ugm").
			JoinOn("ugm.group_id = \"group\".group_id").
			Where("\"group\".workflow_status = ?", common.WorkflowStatusApproved.GetStatusID()).
			Where("ugm.user_id = ?", userID).
			Order("group_name ASC")
	} else {
		query = query.ColumnExpr(`"group".group_id, "group".group_name`).
			Where("\"group\".workflow_status = ?", common.WorkflowStatusApproved.GetStatusID()).
			Order("group_name ASC")
	}
	err = query.Select()

	return groups, err
}

// GetGroupDetailUsingID returns the Group model stored in the database for the given Group ID
func GetGroupDetailUsingID(groupID int64) (*Group, error) {
	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		return nil, fmt.Errorf("Unable to get database client. Err: %s", err.Error())
	}

	group := &Group{
		GroupID: groupID,
	}
	err = client.GetPGClient().Select(group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

// GetAllPresentGroupsKeyVal returns list of all approved groups
func GetAllPresentGroupsKeyVal(log logger.MultiLogger) ([]*GroupKeyVal, error) {
	var groups []*GroupKeyVal

	// Get Database client
	client, err := database.GetClient()
	if err != nil {
		log.Errorf("Unable to get database client. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to get the groups available")
	}
	query := client.GetPGClient().Model(&groups)
	query = query.ColumnExpr(`"group".group_id, "group".group_name`).
		Where("\"group\".workflow_status = ?", common.WorkflowStatusApproved.GetStatusID()).
		Order("group_name ASC")
	err = query.Select()
	if err != nil {
		log.Errorf("Unable to get list of groups. Err: %s", err.Error())
		return nil, fmt.Errorf("Unable to get the groups available")
	}

	return groups, nil
}
