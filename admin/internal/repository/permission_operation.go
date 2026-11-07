package repository

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/database/gdb"
)

func insertPermissionOperationLog(tx gdb.TX, item *model.PermissionOperationLog) error {
	_, err := tx.Model("sys_permission_oper_log").Data(item).Insert()
	return err
}
