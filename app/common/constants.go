package common

// AppRole is the enum for application roles supported
type AppRole string

// WorkflowStatus is the enum for Admin Workflow status representation
type WorkflowStatus int

// ItemType is the enum for ItemTypes supported in the application
type ItemType int

const (

	/*
		APPLICATION ROLES
	*/

	// RoleAdmin is the applications admin
	RoleAdmin AppRole = "Admin"
	// RoleGroupAdmin handles a particular group
	RoleGroupAdmin AppRole = "Group Admin"
	// RoleMember is a member of one or more groups
	RoleMember AppRole = "Member"

	/*
		WORKFLOW STATUS IDs
	*/

	// WorkflowStatusPending signifies that the record is pending for admin's action
	// The object is unusable till it is approved
	WorkflowStatusPending WorkflowStatus = 0
	// WorkflowStatusApproved signifies the record is approved by admin
	WorkflowStatusApproved WorkflowStatus = 1
	// WorkflowStatusRejected signifies the record is rejected by admin
	// The object cannot be used further
	WorkflowStatusRejected WorkflowStatus = 2

	/*
		ITEM TYPES
	*/

	// ItemTypeUnknown represents the items with type unknown
	ItemTypeUnknown ItemType = 0

	// ItemTypePictures represents the items of type Pictures
	ItemTypePictures ItemType = 1

	// ItemTypeVideos represents the items of type Videos
	ItemTypeVideos ItemType = 2

	/*
		LIMITS
	*/

	// UserMaxAllowedItems is the limit in number of items a user can upload in the app
	UserMaxAllowedItems = 20
	// UserMaxAllowedSpace is the limit in max space the user is allocated in the app
	UserMaxAllowedSpace = 100.0

	// GroupMaxAllowedItems is the limit in number of items a user can upload in the app
	GroupMaxAllowedItems = 100
	// GroupMaxAllowedSpace is the limit in max space the user is allocated in the app
	GroupMaxAllowedSpace = 500.0
)

// GetString returns string representation of workflow status
func (w WorkflowStatus) GetString() string {
	switch w {
	case WorkflowStatusPending:
		return "Pending for approval"
	case WorkflowStatusApproved:
		return "Approved"
	case WorkflowStatusRejected:
		return "Rejected"
	}

	return ""
}

// GetStatusID returns integer status value associated with workflow status
func (w WorkflowStatus) GetStatusID() int {
	return int(w)
}

// GetString returns string representation of the item type
func (i ItemType) GetString() string {
	switch i {
	case ItemTypePictures:
		return "Picture"
	case ItemTypeVideos:
		return "Video"
	}

	return "Unknown"
}

// GetItemID returns integer status value associated with ItemType enum
func (i ItemType) GetItemID() int {
	return int(i)
}
