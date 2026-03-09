package blacklist

import (
	"fmt"
	"sync"
)

var blackList sync.Map

func init() {
	blackList = sync.Map{}
}

func userId2Key(id string) string {
	return fmt.Sprintf("userid_%s", id)
}

func BanUser(id string) {
	blackList.Store(userId2Key(id), true)
}

func UnbanUser(id string) {
	blackList.Delete(userId2Key(id))
}

func IsUserBanned(id string) bool {
	_, ok := blackList.Load(userId2Key(id))
	return ok
}
