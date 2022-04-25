package chess

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

/*
Game 象棋窗口==结构体
*号用于指定变量是作为一个指针。
*/
type Game struct {
	sqSelected     int                   //选中的格子
	mvLast         int                   //上一步棋
	bFlipped       bool                  //是否翻转棋盘
	bGameOver      bool                  //是否游戏结束
	showValue      string                //显示内容
	images         map[int]*ebiten.Image //图片资源
	audios         map[int]*audio.Player //音效
	audioContext   *audio.Context        //音效器
	singlePosition *Position             //棋局单列
}

//创建象棋程序
func NewGame() bool {
	//Go 语言的取地址符是 &，放到一个变量前使用就会返回相应变量的内存地址。
	game := &Game{
		images:         make(map[int]*ebiten.Image),
		audios:         make(map[int]*audio.Player),
		singlePosition: NewPosition(),
	}
	if game == nil || game.singlePosition == nil {
		return false
	}
	//初始化audioContext //音效器
	game.audioContext = audio.NewContext(sampleRate)
	//加载资源
	if ok := game.loadResource(); !ok {
		return false
	}
	//加载开局库
	game.singlePosition.loadBook()
	game.singlePosition.startup()

	//设置窗口，接收信息
	ebiten.SetWindowSize(BoardWidth, BoardHeight)
	ebiten.SetWindowTitle("中国象棋")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

//加载资源
func (g *Game) loadResource() bool {
	l := &Load{}
	for k, v := range resMap {
		if k >= MusicSelect {
			//加载音效
			player, err := l.LoadWav(v, g.audioContext)
			if err != nil {
				return false
			}
			g.audios[k] = player
		} else {
			//加载图片
			img := l.LoadImage(v)
			if img == nil {
				return false
			}
			g.images[k] = img
		}
	}
	return true
}

//更新状态，1秒60帧
////该 method 属于 Game 类型对象中的方法
func (g *Game) Update() error {

	return nil
}

//绘制屏幕
func (g *Game) Draw(screen *ebiten.Image) {

	g.drawBoard(screen)
	//IsMouseButtonJustPressed坚持对应的按钮有没有触发
	//MouseButtonLeft表示鼠标左键，象棋游戏只需要用到鼠标左键
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.bGameOver {
			g.bGameOver = false
			g.showValue = ""
			g.sqSelected = 0
			g.mvLast = 0
			//g.singlePosition.startup()
		} else {
			x, y := ebiten.CursorPosition()
			x = Left + (x-BoardEdge)/SquareSize
			y = Top + (y-BoardEdge)/SquareSize
			g.clickSquare(screen, x, y)
		}
	}

	if g.bGameOver {
		g.messageBox(screen)
	}
}

//布局采用外部尺寸并返回（逻辑）屏幕尺寸，如果不使用外部尺寸，只需要返回固定尺寸即可。
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return BoardWidth, BoardHeight
}

//播放音效
func (g *Game) playAudio(value int) bool {
	if player, ok := g.audios[value]; ok {
		if err := player.Rewind(); err != nil {
			return false
		}
		player.Play()
		return true
	}
	return false
}

//绘制棋盘,并且加载棋子的位置
func (g *Game) drawBoard(screen *ebiten.Image) {
	//棋盘
	if v, ok := g.images[ImgChessBoard]; ok {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 0)
		screen.DrawImage(v, op)
	}
	//棋子
	for x := Left; x <= Right; x++ {
		for y := Top; y <= Bottom; y++ {
			xPos, yPos := 0, 0 //初始化,图片x、y坐标
			if g.bFlipped {
				xPos = BoardEdge + (xFlip(x)-Left)*SquareSize
				yPos = BoardEdge + (yFlip(y)-Top)*SquareSize
			} else {
				xPos = BoardEdge + (x-Left)*SquareSize
				yPos = BoardEdge + (y-Top)*SquareSize
			}
			sq := squareXY(x, y) ////棋子所在棋盘的纵坐标,值范围[0-255]
			pc := g.singlePosition.ucpcSquares[sq]

			if pc != 0 {
				g.drawPiece(xPos, yPos+5, screen, g.images[pc]) //绘制棋子
			}

			if sq == g.sqSelected {
				const scaleParam = 1.02
				xScale := float64(xPos) / scaleParam
				yScale := float64(yPos) / scaleParam
				g.drawPieceScale(xScale, yScale, scaleParam, scaleParam, screen, g.images[pc])
			}

			if sq == src(g.mvLast) || sq == dst(g.mvLast) {
				g.drawPiece(xPos, yPos, screen, g.images[ImgSelect]) //绘制棋子
			}
		}
	}

}

//绘制棋子
func (g *Game) drawPiece(x, y int, screen, img *ebiten.Image) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

//绘制缩放棋子
func (g *Game) drawPieceScale(x, y, scaleX, scaleY float64, screen, img *ebiten.Image) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(scaleX, scaleY)
	screen.DrawImage(img, op)
}

//messageBox 提示
/*
truetype.Parse解析字体
truetype.NewFace初始化字体
text.Draw在屏幕上显示内容
*/
func (g *Game) messageBox(screen *ebiten.Image) {
	fmt.Println(g.showValue) //.ArcadeN_ttf
	tt, err := truetype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		fmt.Print(err)
		return
	}
	arcadeFont := truetype.NewFace(tt, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})

	text.Draw(screen, g.showValue, arcadeFont, 180, 288, color.White)
	text.Draw(screen, "Click mouse to restart", arcadeFont, 100, 320, color.White)
}

//aiMove AI移动
func (g *Game) aiMove(screen *ebiten.Image) {
	//AI走一步棋
	g.singlePosition.searchMain()
	g.singlePosition.makeMove(g.singlePosition.search.mvResult)
	//把AI走的棋标记出来
	g.mvLast = g.singlePosition.search.mvResult
	//检查重复局面
	vlRep := g.singlePosition.repStatus(3)
	if g.singlePosition.isMate() {
		//如果分出胜负，那么播放胜负的声音
		g.playAudio(MusicGameWin)
		g.showValue = "Your Lose!"
		g.bGameOver = true
	} else if vlRep > 0 {
		vlRep = g.singlePosition.repValue(vlRep)
		//vlRep是对玩家来说的分值
		if vlRep < -WinValue {
			g.playAudio(MusicGameLose)
			g.showValue = "Your Lose!"
		} else {
			if vlRep > WinValue {
				g.playAudio(MusicGameWin)
				g.showValue = "Your Lose!"
			} else {
				g.playAudio(MusicGameWin)
				g.showValue = "Your Draw!"
			}
		}
		g.bGameOver = true
	} else if g.singlePosition.nMoveNum > 100 {
		g.playAudio(MusicGameWin)
		g.showValue = "Your Draw!"
		g.bGameOver = true
	} else {
		//如果没有分出胜负，那么播放将军、吃子或一般走子的声音
		if g.singlePosition.inCheck() {
			g.playAudio(MusicJiang)
		} else {
			if g.singlePosition.captured() {
				g.playAudio(MusicEat)
			} else {
				g.playAudio(MusicPut)
			}
		}
		if g.singlePosition.captured() {
			g.singlePosition.setIrrev()
		}
	}
}

//clickSquare 点击格子处理
func (g *Game) clickSquare(screen *ebiten.Image, x, y int) {
	sq := squareXY(x, y)
	pc := 0
	if g.bFlipped {
		pc = g.singlePosition.ucpcSquares[squareFlip(sq)]

	} else {
		pc = g.singlePosition.ucpcSquares[sq]
	}

	//按位与运算符"&"是双目运算符。 其功能是参与运算的两数各对应的二进位相与
	if (pc & sideTag(g.singlePosition.sdPlayer)) != 0 { //值为（8、16、0）
		//如果点击自己的棋子，那么直接选中
		g.sqSelected = sq
		g.playAudio(MusicSelect)
	} else if g.sqSelected != 0 && !g.bGameOver {
		//如果点击的不是自己的棋子，但有棋子选中了(一定是自己的棋子)，那么走这个棋子
		mv := move(g.sqSelected, sq)
		if g.singlePosition.legalMove(mv) {
			if g.singlePosition.makeMove(mv) {
				g.mvLast = mv
				g.sqSelected = 0
				// 检查重复局面
				vlRep := g.singlePosition.repStatus(3)
				if g.singlePosition.isMate() {
					// 如果分出胜负，那么播放胜负的声音，并且弹出不带声音的提示框
					g.playAudio(MusicGameWin)
					g.showValue = "Your Win!"
					g.bGameOver = true
				} else if vlRep > 0 {
					vlRep = g.singlePosition.repValue(vlRep)
					if vlRep > WinValue {
						g.playAudio(MusicGameLose)
						g.showValue = "Your Lose!"
					} else {
						if vlRep < -WinValue {
							g.playAudio(MusicGameWin)
							g.showValue = "Your Win!"
						} else {
							g.playAudio(MusicGameWin)
							g.showValue = "Your Draw!"
						}
					}
					g.bGameOver = true
				} else if g.singlePosition.nMoveNum > 100 {
					g.playAudio(MusicGameWin)
					g.showValue = "Your Draw!"
					g.bGameOver = true
				} else {
					if g.singlePosition.checked() {
						g.playAudio(MusicJiang)
					} else {
						if g.singlePosition.captured() {
							g.playAudio(MusicEat)
							g.singlePosition.setIrrev()
						} else {
							g.playAudio(MusicPut)
						}
					}

					time.Sleep(1 * time.Second)
					g.aiMove(screen)
				}
			} else {
				g.playAudio(MusicJiang) // 播放被将军的声音
			}
		}
		//如果根本就不符合走法(例如马不走日字)，那么不做任何处理
	}
}
