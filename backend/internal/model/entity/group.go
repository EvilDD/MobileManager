// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Group is the golang structure for table group.
type Group struct {
	Id          int         `json:"id"          orm:"id"          ` //
	Name        string      `json:"name"        orm:"name"        ` //
	Description string      `json:"description" orm:"description" ` //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  ` //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  ` //
}
