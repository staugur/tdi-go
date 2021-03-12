package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "构建文档",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.Flags()
		hour, err := flagset.GetUint("hour")
		if err != nil {
			fmt.Printf("invalid param(hour): %v\n", hour)
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().UintP("hour", "", 12, "Expiration clearance limit")
}
