package commands

import (
	"gocli/internal/inputmethod"
	"strconv"

	"github.com/spf13/cobra"
)

var SetTermCmd = &cobra.Command{
	Use:   "word [word] [pinyin] [weight]",
	Short: "设置输入法词条",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		word := args[0]
		pinyin := args[1]
		weight := 0
		if len(args) >= 3 {
			weight, _ = strconv.Atoi(args[2])
		}

		if word == "" || pinyin == "" {
			cmd.PrintErrln("错误: word 和 pinyin 不能为空")
			return
		}

		term := &inputmethod.WordEntry{
			Word:   word,
			Pinyin: pinyin,
			Weight: weight,
		}

		if err := inputmethod.SetTerm(term); err != nil {
			cmd.PrintErrln("设置词条失败:", err)
			return
		}

		cmd.Println("成功添加/更新词条:", term.Word)
	},
}

var RestartInputMethodCmd = &cobra.Command{
	Use:   "rime restart",
	Short: "重启输入法服务",
	Run: func(cmd *cobra.Command, args []string) {
		if err := inputmethod.RestartInputMethod(); err != nil {
			cmd.PrintErrln("重启输入法服务失败:", err)
			return
		}
		cmd.Println("输入法服务重启成功")
	},
}
