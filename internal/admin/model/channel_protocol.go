package model

const (
	channelProtocolTableName = "channel_protocol"
)

// ChannelProtocolCatalog stores channel protocol options for admin UI.
// `name` is the protocol key and should be treated as the canonical identifier.
type ChannelProtocolCatalog struct {
	Name        string `json:"name" gorm:"type:varchar(64);not null;uniqueIndex:idx_channel_protocol_name"`
	ProtocolID  int    `json:"protocol_id" gorm:"column:id;index"`
	Label       string `json:"label" gorm:"type:varchar(128);default:''"`
	Color       string `json:"color" gorm:"type:varchar(32);default:''"`
	Description string `json:"description" gorm:"type:text"`
	Tip         string `json:"tip" gorm:"type:text"`
	Source      string `json:"source" gorm:"type:varchar(32);default:'default'"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	SortOrder   int    `json:"sort_order" gorm:"column:sort_order;default:0"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

func (ChannelProtocolCatalog) TableName() string {
	return channelProtocolTableName
}
