package model

import "context"

type GroupModelRouteRepository struct {
	GetRandomSatisfiedChannel   func(group string, model string, ignoreFirstPriority bool) (*Channel, error)
	ListSatisfiedChannels       func(group string, model string) ([]*Channel, error)
	AddGroupModelRoutes         func(channel *Channel) error
	DeleteGroupModelRoutes      func(channel *Channel) error
	UpdateGroupModelRoutes      func(channel *Channel) error
	UpdateGroupModelRouteStatus func(channelId string, status bool) error
	GetTopChannelByModel        func(group string, model string) (*Channel, error)
	GetGroupModels              func(ctx context.Context, group string) ([]string, error)
}

var groupModelRouteRepo GroupModelRouteRepository

func BindGroupModelRouteRepository(repo GroupModelRouteRepository) {
	groupModelRouteRepo = repo
}

func mustGroupModelRouteRepo() GroupModelRouteRepository {
	if groupModelRouteRepo.GetRandomSatisfiedChannel == nil {
		panic("group model route runtime repository not initialized")
	}
	if groupModelRouteRepo.ListSatisfiedChannels == nil {
		panic("group model route runtime repository not initialized")
	}
	return groupModelRouteRepo
}
