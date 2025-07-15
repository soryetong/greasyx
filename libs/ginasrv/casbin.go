package ginasrv

import (
	"context"
	"errors"

	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/ginahelper"
)

type CasbinInfo struct {
	Path   string `json:"path"`   // 路径
	Method string `json:"method"` // 方法
}

func UpsertCasbin(ctx context.Context, roleId int64, casbinInfos []CasbinInfo) error {
	id := ginahelper.Int64ToString(roleId)
	clearCasbin(0, id)
	rules := [][]string{}
	for _, v := range casbinInfos {
		rules = append(rules, []string{id, v.Path, v.Method})
	}

	success, err := gina.Casbin.AddPolicies(rules)
	if !success {
		return errors.New("存在相同api,添加失败,请联系管理员")
	}

	return err
}

func clearCasbin(v int, p ...string) bool {
	success, _ := gina.Casbin.RemoveFilteredPolicy(v, p...)
	return success
}
