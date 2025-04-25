package model

import "github.com/gogf/gf/v2/os/gtime"

// File 文件模型
type File struct {
	Id           int64       `json:"id"           description:"文件ID"`
	Name         string      `json:"name"         description:"文件名称"`
	OriginalName string      `json:"originalName" description:"原始文件名"`
	FileType     string      `json:"fileType"     description:"文件类型"`
	FileSize     int64       `json:"fileSize"     description:"文件大小"`
	FilePath     string      `json:"filePath"     description:"文件路径"`
	MimeType     string      `json:"mimeType"     description:"MIME类型"`
	Md5          string      `json:"md5"          description:"MD5值"`
	Status       int         `json:"status"       description:"状态 1:正常 0:删除"`
	CreatedAt    *gtime.Time `json:"createdAt"    description:"创建时间"`
	UpdatedAt    *gtime.Time `json:"updatedAt"    description:"更新时间"`
}
