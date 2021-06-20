package main

// 使用命令行ui库来绘制，避免闪烁
import (
	"os"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func startMenu() {
	err := ui.Init()
	handleErr(err, "CUI初始化")
	defer ui.Close()
	termWidth, termHeight := ui.TerminalDimensions()
	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)
	optionList := widgets.NewList()
	optionList.Title = "Options"
	optionList.Rows = []string{
		"[0] start 开始程序",
		"[1] exit",
		"[2] continue",
	}
	voidBlock := ui.NewBlock()
	voidBlock.Border = false
	grid.Set(ui.NewRow(1.0/2, ui.NewCol(1.0, voidBlock)), ui.NewRow(1.0/2, ui.NewCol(0.25, voidBlock), ui.NewCol(0.5, optionList), ui.NewCol(0.25, voidBlock)))
	ui.Render(grid)

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
func drawcui() {
	err := ui.Init()
	handleErr(err, "CUI初始化")
	defer ui.Close()

	statusList := widgets.NewList()
	statusList.Title = "Status"
	statusList.Rows = []string{
		"[0] github.com/gizak/termui/v3",
		"[1] [你好，世界]",
		"[2] [こんにちは世界",
		"[3] [color](fg:white,bg:green) output",
		"[4] output.go",
		"[5] random_out.go",
		"[6] dashboard.go",
		"[7] foo",
		"[8] bar",
		"[9] baz",
	}
	statusList.TextStyle = ui.NewStyle(ui.ColorYellow)
	statusList.WrapText = false
	statusList.SetRect(0, 0, 100, 100)
	ui.Render(statusList)
	os.Exit(0)
	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
}
