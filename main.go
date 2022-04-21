package main

import (
	"HenanChess/chess"
	//"HenanChess/res"
	"fmt"
)

func main() {

	// if err := res.FileToByte("./res", "./chess"); err != nil {
	// 	fmt.Println("文件转换失败！")
	// }

	if ok := chess.NewGame(); !ok {
		fmt.Println("游戏启动失败！")
	}
}
