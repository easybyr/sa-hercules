package repository

import "strconv"

func directRoleCode(uid int64) string { return "user_" + strconv.FormatInt(uid, 10) + "_direct" }
