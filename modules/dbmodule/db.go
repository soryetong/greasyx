package dbmodule

import (
	"encoding/json"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	console.Append(dbCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Init DB",
	Long:  `加载DB模块`,
	Run: func(cmd *cobra.Command, args []string) {
		initFunc()
	},
}

type dbConfig struct {
	Dsn             string
	Driver          string
	UseOrm          bool
	LogLevel        int
	EnableLogWriter bool
	MaxIdleConn     int
	MaxConn         int
	SlowThreshold   int
}

func initFunc() {
	conf := viper.Get(`Db`)
	confMap, ok := conf.([]interface{})
	if !ok || len(confMap) == 0 {
		console.Echo.Fatalf("❌ 错误: 请确保 `Db` 模块的配置符合要求\n")
	}

	initMap()

	for _, v := range confMap {
		dbConfMap, ok := v.(map[string]interface{})
		if !ok {
			console.Echo.Errorf("❌ 类型断言错误: 请确保 `Db` 模块的配置符合要求\n")
			continue
		}

		jsonData, err := json.Marshal(dbConfMap)
		if err != nil {
			console.Echo.Errorf("❌ json.Marshal错误: 请确保 `Db` 模块的配置符合要求\n")
			continue
		}

		var dbConf dbConfig
		if err = json.Unmarshal(jsonData, &dbConf); err != nil {
			console.Echo.Errorf("❌ json.Unmarshal错误: 请确保 `Db` 模块的配置符合要求\n")
			continue
		}

		if dbConf.Driver == "" || dbConf.Dsn == "" {
			console.Echo.Fatalf("❌ 错误: 你正在加载Db模块，但是你未配置Dsn和Driver，请先添加配置\n")
			continue
		}

		if dbConf.UseOrm {
			initGorm(&dbConf)
		} else {
			initSqlx(&dbConf)
		}
	}
}

func initMap() {
	om[gina.DbTypeMysql] = "gina.GMySQL()"
	om[gina.DbTypePostgresql] = "gina.GPostgres()"
	om[gina.DbTypeSqlite] = "gina.GSqlite()"
	om[gina.DbTypeSqlserver] = "gina.GSqlserver()"
	om[gina.DbTypeOracle] = ""

	sm[gina.DbTypeMysql] = "gina.XMySQL()"
	sm[gina.DbTypePostgresql] = "gina.XPostgres()"
	sm[gina.DbTypeSqlite] = "gina.XSqlite()"
	sm[gina.DbTypeSqlserver] = "gina.XSqlserver()"
	sm[gina.DbTypeOracle] = "gina.XOracle()"
}
