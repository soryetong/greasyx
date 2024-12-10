package mongomodule

import (
	"context"
	"fmt"
	"github.com/soryetong/greasyx/gina"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func init() {
	gina.Append(mongoCmd)
}

var mongoCmd = &cobra.Command{
	Use:   "MongoDB",
	Short: "Init MongoDB",
	Long:  `加载MongoDB模块之后，可以通过 gina.Mdb 进行数据操作`,
	Run: func(cmd *cobra.Command, args []string) {
		url := viper.GetString("Mongo.Url")
		if url == "" {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] "+
				"你正在加载MongoDB模块，但是你未配置Mongo.Url，请先添加配置\n")
			os.Exit(124)
		}

		initClient(url)
	},
}

func initClient(url string) {
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-ERROR] "+
			"MongoDB连接失败: %s\n", err))
		os.Exit(124)
	}

	// 检查连接
	if err = client.Ping(context.TODO(), nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-ERROR] "+
			"MongoDB连接失败: %s\n", err))
		os.Exit(124)
	}

	gina.Mdb = client
	_, _ = fmt.Fprintf(os.Stderr, "\n\033[32m [GREASYX-GINFO] "+
		"Mongo模块加载成功, 你可以使用 `gina.Mdb` 进行数据操作\033[0m\n")
}
