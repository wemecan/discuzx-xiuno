package controllers

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/g/database/gdb"
	"time"
	"xiuno-tools/app/libraries/common"
	"xiuno-tools/app/libraries/database"
	"xiuno-tools/app/libraries/mcfg"
	"xiuno-tools/app/libraries/mlog"
)

type user struct {
}

func (t *user) ToConvert() (err error) {
	start := time.Now()

	cfg := mcfg.GetCfg()

	//ucPre := cfg.GetString("database.uc.0.prefix")
	//discuzPre := cfg.GetString("database.discuz.0.prefix")
	//xiunoPre := cfg.GetString("database.xiuno.0.prefix")

	ucPre, discuzPre, xiunoPre := database.GetPrefix("uc"), database.GetPrefix("discuz"), database.GetPrefix("xiuno")

	ucMemberTable := ucPre + "members"
	dxMemberTable := discuzPre + "common_member"
	dxMemberStatusTable := discuzPre + "common_member_status"

	fields := "m.uid,m.groupid,m.email,m.username,m.credits,m.regdate,s.regip,s.lastip,s.lastvisit,u.password,u.salt"
	var r gdb.Result
	r, err = database.GetDiscuzDB().Table(dxMemberTable+" m").LeftJoin(dxMemberStatusTable+" s", "s.uid = m.uid").LeftJoin(ucMemberTable+" u", "u.uid = m.uid").Fields(fields).Select()

	xiunoTable := xiunoPre + cfg.GetString("tables.xiuno.user.name")
	if err != nil {
		mlog.Log.Debug("", "表 %s 数据查询失败, %s", xiunoTable, err.Error())
	}

	if len(r) == 0 {
		mlog.Log.Debug("", "表 %s 无数据可以转换", xiunoTable)
		return nil
	}

	xiunoDB := database.GetXiunoDB()
	if _, err = xiunoDB.Exec("TRUNCATE " + xiunoTable); err != nil {
		return errors.New(fmt.Sprintf("清空数据表(%s)失败, %s", xiunoTable, err.Error()))
	}

	dataList := gdb.List{}
	for _, u := range r.ToList() {
		dataList = append(dataList, gdb.Map{
			"uid":         u["uid"],
			"gid":         u["groupid"],
			"email":       u["email"],
			"username":    u["username"],
			"password":    u["password"],
			"salt":        u["salt"],
			"credits":     u["credits"],
			"create_ip":   common.Ip2long(u["regip"].(string)),
			"create_date": u["create_date"],
			"login_ip":    common.Ip2long(u["lastip"].(string)),
			"login_date":  u["lastvisit"],
		})
	}

	if res, err := xiunoDB.BatchInsert(xiunoTable, dataList, 100); err != nil {
		return errors.New(fmt.Sprintf("表 %s 数据插入失败, %s", xiunoTable, err.Error()))
	} else {
		count, _ := res.RowsAffected()
		mlog.Log.Info("", fmt.Sprintf("表 %s 数据导入成功, 本次导入: %d 条数据, 耗时: %v", xiunoTable, count, time.Since(start)))
		return nil
	}
}

func NewUser() *user {
	t := &user{}
	return t
}