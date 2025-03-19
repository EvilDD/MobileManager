// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Device is the golang structure for table device.
type Device struct {
	Id        int         `json:"id"        orm:"id"         ` //
	Name      string      `json:"name"      orm:"name"       ` //
	DeviceId  string      `json:"deviceId"  orm:"device_id"  ` //
	Status    string      `json:"status"    orm:"status"     ` //
	GroupId   int         `json:"groupId"   orm:"group_id"   ` //
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" ` //
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" ` //
}
