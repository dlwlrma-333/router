package model

// ModelProvider stores model provider catalog in a dedicated table.
type ModelProvider struct {
	Id        string `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	Name      string `json:"name" gorm:"type:varchar(128);default:''"`
	BaseURL   string `json:"base_url" gorm:"column:base_url;type:text"`
	SortOrder int    `json:"sort_order" gorm:"column:sort_order;type:int;not null;default:1000"`
	Source    string `json:"source" gorm:"type:varchar(32);default:'manual'"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint"`
}

func (ModelProvider) TableName() string {
	return "providers"
}
