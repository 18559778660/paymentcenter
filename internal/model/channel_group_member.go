package model

// ChannelGroupMember 通道分组与账号的多对多关系。
type ChannelGroupMember struct {
	GroupID          uint `gorm:"column:group_id;primaryKey;comment:分组ID" json:"group_id"`
	ChannelAccountID uint `gorm:"column:channel_account_id;primaryKey;index;comment:通道账号ID" json:"channel_account_id"`
}

func (ChannelGroupMember) TableName() string {
	return "channel_group_members"
}
