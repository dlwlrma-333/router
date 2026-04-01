package plan

import "github.com/yeying-community/router/internal/admin/model"

func ListPage(page int, pageSize int, keyword string) ([]model.ServicePackage, int64, error) {
	return model.ListServicePackagesPage(page, pageSize, keyword)
}

func Get(id string) (model.ServicePackage, error) {
	return model.GetServicePackageByID(id)
}

func Create(item model.ServicePackage) (model.ServicePackage, error) {
	return model.CreateServicePackage(item)
}

func Update(item model.ServicePackage) (model.ServicePackage, error) {
	return model.UpdateServicePackage(item)
}

func Delete(id string) error {
	return model.DeleteServicePackage(id)
}

func AssignToUser(id string, userID string, startAt int64) (model.UserPackageSubscription, error) {
	return model.AssignServicePackageToUser(id, userID, startAt)
}
