package srv

import (
	"qipai/dao"
	"qipai/model"
	"testing"
)

func TestRoom(t *testing.T) {
	var ids []int
	dao.Db.Raw("select id  from rooms  where deleted_at is null and club_id=0 and now()-created_at>10").Pluck("id",&ids)
	dao.Db.Unscoped().Where("id in (?)", ids).Delete(model.Room{})
}