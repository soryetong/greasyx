package mongomodule

import (
	"context"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	console.Append(mongoCmd)
}

var mongoCmd = &cobra.Command{
	Use:   "MongoDB",
	Short: "Init MongoDB",
	Long:  `加载MongoDB模块之后，可以通过 gina.Mdb 进行数据操作`,
	Run: func(cmd *cobra.Command, args []string) {
		url := viper.GetString("Mongo.Url")
		if url == "" {
			console.Echo.Fatalln("❌ 错误: 你正在加载MongoDB模块，但是你未配置Mongo.Url，请先添加配置\n")
		}

		initClient(url)
	},
}

func initClient(url string) {
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		console.Echo.Fatalf("❌ 错误: MongoDB连接失败: %s\n", err)
	}

	// 检查连接
	if err = client.Ping(context.TODO(), nil); err != nil {
		console.Echo.Fatalf("❌ 错误: MongoDB连接失败: %s\n", err)
	}

	gina.Mdb = client
	console.Echo.Info("✅ 提示: Mongo模块加载成功, 你可以使用 `gina.Mdb` 进行数据操作\n")
}
