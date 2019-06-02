package srv

import (
	"errors"
	"fmt"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
	"time"
)

var Club clubSrv

type clubSrv struct {
}

func (this *clubSrv) CreateClub(club *model.Club) (err error) {
	dao.Db().Save(club)
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
	dao.Db().Where("uid = ? and status <> 0", uid).Find(&cus)
	var ids []uint
	for _, v := range cus {
		ids = append(ids, v.ClubId)
	}
	dao.Db().Where(ids).Find(&clubs)

	return
}

func (clubSrv) Join(clubId, userId uint) (err error) {

	// 查询出俱乐部信息
	var club model.Club
	dao.Db().First(&club, clubId)
	if club.ID == 0 {
		err = errors.New(fmt.Sprintf("编号为%d的俱乐部不存在", clubId))
		return
	}

	// 防止重复加入
	var n int
	dao.Db().Model(&model.ClubUser{}).Where(&model.ClubUser{Uid: userId, ClubId: clubId}).Count(&n)
	if n > 0 {
		err = errors.New("您已经是该俱乐部会员")
		return
	}

	cu := &model.ClubUser{
		Uid:    userId,
		ClubId: clubId,
	}

	// 如果俱乐部不需要审核，用户就直接成为正式用户
	// 如果要加入的用户正好是老板，直接成为正式用户
	if !club.Check || club.Uid == userId {
		cu.Status = enum.ClubUserVip
	}

	dao.Db().Save(cu)
	return
}

func (clubSrv) UpdateInfo(clubId uint, check, close bool, name, rollText, notice string) (err error) {
	var club model.Club
	dao.Db().First(&club, clubId)
	if club.ID == 0 {
		err = errors.New("该俱乐部不存在")
		return
	}
	club.Check = check
	club.Close = close
	club.Name = name
	club.Notice = notice
	club.RollText = rollText
	dao.Db().Save(&club)
	return
}

func (this *clubSrv) IsClubUser(userId, clubId uint) (ok bool) {
	var n int
	dao.Db().Model(&model.ClubUser{}).Where(&model.ClubUser{Uid: userId, ClubId: clubId}).Count(&n)
	ok = n > 0
	return
}

type ClubUser struct {
	Id        uint              `json:"id"`
	Nick      string            `json:"nick"`
	Avatar    string            `json:"avatar"`
	ClubId    uint              `json:"club_id"` // 俱乐部编号
	Status    enum.ClubUserType `json:"status"`  // 0 等待审核，1 正式用户， 2 冻结用户
	Admin     bool              `json:"admin"`   // 是否是管理员 true 是管理员
	CreatedAt time.Time         `json:"created_at"`
	DeletedAt *time.Time        `json:"deleted_at"`
}

func (this *clubSrv) Users(clubId uint) (users []ClubUser) {

	dao.Db().Table("club_users").
		Select("users.id,users.nick, users.avatar,club_users.admin,club_users.status,club_users.created_at,club_users.deleted_at").
		Joins("join users on club_users.uid=users.id").Where("club_users.club_id = ?", clubId).Scan(&users)

	//var cus []model.ClubUser
	//var ids []uint
	//dao.Db().Where(&model.ClubUser{ClubId: clubId}).Find(&cus)
	//for _, v := range cus {
	//	ids = append(ids, v.Uid)
	//}
	//dao.Db().Where(ids).Find(&users)
	return
}

func (this *clubSrv) getClubUser(clubId, userId uint) (cu model.ClubUser, err error) {
	if !this.IsClubUser(userId, clubId) {
		err = errors.New("用户不属于该俱乐部")
		return
	}
	dao.Db().Where(&model.ClubUser{ClubId: clubId, Uid: userId}).First(&cu)
	if cu.ID == 0 {
		err = errors.New("没在俱乐部找到该用户")
		return
	}

	return
}

// 设置、取消管理
func (this *clubSrv) SetAdmin(clubId, userId uint, ok bool) (err error) {
	var cu model.ClubUser
	cu, err = this.getClubUser(clubId, userId)
	if err != nil {
		return
	}
	if ok && cu.Admin {
		err = errors.New("该用户已经是管理员")
		return
	}
	cu.Admin = ok
	dao.Db().Save(&cu)
	return
}

// 冻结、取消冻结,ok 为true表示冻结
func (this *clubSrv) SetDisable(clubId, userId uint, ok bool) (err error) {
	var cu model.ClubUser
	cu, err = this.getClubUser(clubId, userId)
	if err != nil {
		return
	}

	if ok {
		if cu.Status != enum.ClubUserVip {
			err = errors.New("该用户还不是正式会员，无法冻结")
			return
		}
		cu.Status = enum.ClubUserDisable
	} else {
		cu.Status = enum.ClubUserVip
	}
	dao.Db().Save(&cu)
	return
}

// 设置、取消代付
func (this *clubSrv) SetPay(clubId, userId uint, ok bool) (err error) {
	var cu model.ClubUser
	cu, err = this.getClubUser(clubId, userId)
	if err != nil {
		return
	}
	var club model.Club
	dao.Db().First(&club, cu.ClubId)
	if club.ID == 0 {
		err = errors.New("没找到该俱乐部")
		return
	}

	// 如果是取消代付，先判断当前用户是否是代付者
	if !ok && userId != club.PayerUid {
		err = errors.New("该账号不是代付账号")
		return
	}

	if ok {
		// 设置代付
		club.PayerUid = userId
	} else {
		// 取消代付
		club.PayerUid = 0
	}

	dao.Db().Save(&club)
	return
}

// 移除用户
func (this *clubSrv) RemoveClubUser(clubId, userId uint) (err error) {
	_, err = this.getClubUser(clubId, userId)
	if err != nil {
		return
	}

	// 如果是代付，无法直接删除
	var club model.Club
	dao.Db().First(&club, clubId)
	if club.ID == 0 {
		err = errors.New("该俱乐部不存在")
		return
	}

	if club.PayerUid != 0 && userId == club.PayerUid {
		err = errors.New("该用户是代付，请先取消代付之后再删除")
		return
	}

	dao.Db().Unscoped().Where("club_id=? and uid=?", clubId, userId).Delete(model.ClubUser{})
	return
}

// 检查操作人员是不是俱乐部管理员
func (this *clubSrv) IsAdmin(opUid, clubId uint) (ok bool) {
	cu, err := this.getClubUser(clubId, opUid)
	if err != nil {
		return
	}
	ok = cu.Admin
	return
}

// 检查操作人员是不是俱乐部老板
func (clubSrv) IsBoss(opUid, clubId uint) (ok bool) {
	var club model.Club
	dao.Db().First(&club, clubId)
	if club.Uid == opUid {
		ok = true
	}
	return
}

// 指定用户获取指定俱乐部
func (this *clubSrv) GetClub(uid, cid uint) (club model.Club, err error) {
	if !this.IsClubUser(uid, cid) {
		err = errors.New("您不是该俱乐部成员")
		return
	}
	dao.Db().First(&club, cid)
	if club.ID == 0 {
		err = errors.New("没找到您指定的俱乐部")
	}
	return
}
