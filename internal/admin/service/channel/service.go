package channel

import (
	"github.com/yeying-community/router/internal/admin/model"
	channelrepo "github.com/yeying-community/router/internal/admin/repository/channel"
)

func GetAll(start, num int, status string) ([]*model.Channel, error) {
	return channelrepo.GetAll(start, num, status)
}

func Search(keyword string) ([]*model.Channel, error) {
	return channelrepo.Search(keyword)
}

func GetByID(id string, selectAll bool) (*model.Channel, error) {
	return channelrepo.GetByID(id, selectAll)
}

func Insert(channel *model.Channel) error {
	return channelrepo.Insert(channel)
}

func DeleteByID(id string) error {
	return channelrepo.DeleteByID(id)
}

func DeleteDisabled() (int64, error) {
	return channelrepo.DeleteDisabled()
}

func Update(channel *model.Channel) error {
	return channelrepo.Update(channel)
}

func UpdateTestModelByID(id string, testModel string) error {
	return channelrepo.UpdateTestModelByID(id, testModel)
}
