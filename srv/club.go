package srv

import (
	"errors"
	"fmt"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
)

var Club clubSrv

type clubSrv struct {
}

func (this *clubSrv) CreateClub(club *model.Club) (err error) {
	dao.Db.Save(club)
	if club.ID == 0 {
		err = errors.New("俱乐部创建失败，请联系管理员")
		return
	}
	// 创建成功后，把自己加入俱乐部
	if err = this.Join(club.ID, club.Uid); err != nil {
		return
	}
	return
}

func (clubSrv) MyClubs(uid uint) (clubs []model.Club) {

	// 我加入的
	var cus []model.ClubUser
	dao.Db.Where(&model.ClubUser{Uid: uid}).Find(&cus)
	var ids []uint
	for _, v := range cus {
		ids = append(ids, v.ClubId)
	}
	dao.Db.Where(ids).Find(&clubs)

	return
}

func (clubSrv) Join(clubId, userId uint) (err error) {

	// 查询出俱乐部信息
	var club model.Club
	dao.Db.First(&club, clubId)
	if club.ID == 0 {
		err = errors.New(fmt.Sprintf("编号为%d的俱乐部不存在", clubId))
		return
	}

	// 防止重复加入
	var n int
	dao.Db.Model(&model.ClubUser{}).Where(&model.ClubUser{Uid: userId, ClubId: clubId}).Count(&n)
	if n > 0 {
		err = errors.New("您已经是该俱乐部会员")
		return
	}

	cu := &model.ClubUser{
		Uid:    userId,
		ClubId: clubId,
	}

	// 如果俱乐部不需要审核，用户就直接成为正式用户
	if !club.Check {
		cu.Status = enum.ClubUserVip
	}

	dao.Db.Save(cu)
	return
}

func (clubSrv) UpdateNameAndNotice(clubId uint, name, notice string) (err error) {
	var club model.Club
	dao.Db.First(&club, clubId)
	if club.ID == 0 {
		err = errors.New("该俱乐部不存在")
		return
	}
	club.Name = name
	club.Notice = notice
	dao.Db.Save(club)
	return
}

func (this *clubSrv) IsClubUser(userId, clubId uint) (ok bool) {
	var n int
	dao.Db.Model(&model.ClubUser{}).Where(&model.ClubUser{Uid: userId, ClubId: clubId}).Count(&n)
	ok = n > 0
	return
}

func (this *clubSrv) Users(clubId uint) (users []model.User) {
	var cus []model.ClubUser
	var ids []uint
	dao.Db.Where(&model.ClubUser{ClubId: clubId}).Find(&cus)
	for _, v := range cus {
		ids = append(ids, v.Uid)
	}
	dao.Db.Where(ids).Find(&users)
	return
}
