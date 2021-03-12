package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tcw.im/ufc"
)

const version = "0.1.0"

var (
	dir    string // download dir
	host   string
	port   uint
	rawurl string // redis connect url
	token  string
	status string
	alert  string // alert email
)

var rootCmd = &cobra.Command{
	Use:   "tdi",
	Short: "花瓣网、堆糖网下载油猴脚本的远程下载服务（Tdi for Go）",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		sv, _ := cmd.Flags().GetBool("version")
		if sv {
			fmt.Println(version)
		} else {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().BoolP(
		"version", "v", false, "show version and exit",
	)

	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", "", "download base directory")
	rootCmd.PersistentFlags().StringVarP(&rawurl,
		"redis", "u", "", "redis url, format: redis://[:<password>@]<host>:<port>/<db>",
	)
}

func initConfig() {
	if dir == "" || !ufc.IsDir(dir) {
		fmt.Println("invalid dir")
		os.Exit(127)
	}
	if rawurl == "" {
		rawurl = os.Getenv("tdi_redis_url")
	}
	if rawurl == "" {
		fmt.Println("invalid redis url")
		os.Exit(127)
	}
}
