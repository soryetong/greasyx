package casbinmodule

import (
	"path/filepath"
	"runtime"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	console.Append(casbinCmd)
}

var casbinCmd = &cobra.Command{
	Use:   "Casbin",
	Short: "Init Casbin",
	Long:  `加载Casbin模块之后，可以通过 gina.Casbin 进行权限校验`,
	Run: func(cmd *cobra.Command, args []string) {
		initCasbin()
	},
}

func initCasbin() {
	modePath := viper.GetString("Casbin.ModePath")
	if modePath == "" {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		modePath = filepath.Join(dir, "rbac_model.conf")
	}

	db := gina.GMySQL()
	if db == nil {
		db = gina.GetGorm(viper.GetString("Casbin.DbName"))
	}
	if db == nil {
		console.Echo.Fatalf("❌ 错误: 你正在加载Casbin模块，但是该模块目前只支持 `MySQL`，请先启用 `gina.GMySQL()`\n")
	}
	a, _ := gormadapter.NewAdapterByDB(db)
	syncedEnforcer, err := casbin.NewSyncedEnforcer(modePath, a)
	if err != nil {
		console.Echo.Fatalf("❌ 错误: Casbin加载失败! %v\n", err)
	}

	_ = syncedEnforcer.LoadPolicy()

	gina.Casbin = syncedEnforcer
	console.Echo.Info("✅ 提示: Casbin模块加载成功, 你可以使用 `gina.Casbin` 进行权限操作\n")
}
