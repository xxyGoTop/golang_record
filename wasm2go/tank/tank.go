package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"syscall/js"
	"time"

	"github.com/Chyroc/web"
)

func timenow() int64 {
	return time.Now().UnixNano() / 100000
}

// 定义一个块
type block struct {
	x float64
	y float64

	width  int
	height int
}

var game *tankGame

// tank game canvas
type tankGame struct {
	canvas web.HTMLCanvasElement
	ctx    web.CanvasRenderingContext2D

	player_lives int //存活
	score        int //分数

	tank_width  int //tank的属性
	tank_height int
	tank_speed  float64

	tank block //tank的属性

	balls           []block //子弹
	ball_speed      float64
	since_last_fire int64

	blocks                    []block //障碍物
	tank_block_collision_bool bool    //是否碰撞

	monsters       []block //障碍物
	monsters_speed float64

	right_pressed bool //按键触发
	left_pressed  bool
	space_pressed bool
}

// 初始化一个tank game
func newTankGame() *tankGame {
	t := tankGame{
		player_lives: 5,
		score:        0,

		tank_width:  30,
		tank_height: 40,
		tank_speed:  0.55,

		ball_speed:      1.6,
		since_last_fire: timenow(),

		tank_block_collision_bool: true,
		monsters_speed:            0.6,
	}

	t.canvas = web.Document.GetElementById("canvas").(web.HTMLCanvasElement)
	t.ctx = t.canvas.GetContext("2d")
	t.tank = block{
		float64(t.canvas.Width()) / 2,
		float64(t.canvas.Height()) - 80,
		t.tank_width,
		t.tank_height,
	}

	return &t
}

// 绘画 tank
func (t *tankGame) drawTank() {
	tank := t.tank

	//绘画 tank 主体
	t.ctx.BeginPath()
	t.ctx.Rect(tank.x, tank.y, tank.width, tank.height)
	t.ctx.SetFillStyle("blue")

	//绘画 tank fire
	t.ctx.Rect(float64(tank.x)+float64(tank.width)/float64(2)-float64(5), tank.y-15, 10, 15)
	t.ctx.SetFillStyle("blue")

	t.ctx.Fill()
	t.ctx.ClosePath()
}

// 绘画 显示栏
func (t *tankGame) drawBorder() {
	canvas := t.canvas

	t.ctx.BeginPath()
	t.ctx.Rect(0, 0, 80, canvas.Height())
	t.ctx.Rect(float64(canvas.Width()-80), 0, 80, canvas.Height())
	t.ctx.SetFillStyle("grey")
	t.ctx.Fill()
	t.ctx.ClosePath()
}

// 初始化一个 new ball
func (t *tankGame) drawNewBall(ball_x, ball_y float64) {
	t.ctx.BeginPath()
	t.ctx.Arc(ball_x, ball_y, 5, 0, math.Pi*2)
	t.balls = append(t.balls, block{ball_x, ball_y, 3, 3})
	t.since_last_fire = timenow()
}

// 绘画所有的ball
func (t *tankGame) drawBalls() {
	for i := 0; i < len(t.balls); i++ {
		t.ctx.BeginPath()
		t.ctx.Arc(t.balls[i].x, t.balls[i].y, 5, 0, math.Pi*2)
		t.ctx.SetFillStyle("red")
		t.ctx.Fill()
		t.ctx.ClosePath()
	}
}

// 生成 随机坐标
func (t *tankGame) generateCoords() []float64 {
	var x = rand.Float64()*float64(t.canvas.Width()-80) + 80
	if int(x) > t.canvas.Width()-80 {
		x = rand.Float64()*float64(t.canvas.Width()-80) + 80
	}

	y := rand.Float64()*float64(-260-60) - 60

	return []float64{x, y}
}

// 检测block 之间的距离
func (t *tankGame) distanceCheck(X1, Y1, X2, Y2 float64) bool {
	distance := math.Sqrt(math.Pow(X1-X2, 2) + math.Pow(Y1-Y2, 2))

	if distance > 140 && math.Abs(Y1-Y2) > 40 {
		return true
	}
	return false
}

// 生成新的block 检测与其他blocks距离
func (t *tankGame) blockDistanceChecker(X, Y float64) bool {
	if len(t.blocks) < 0 {
		return false
	}

	check := false
	for i := 0; i < len(t.blocks); i++ {
		if t.distanceCheck(X, Y, t.blocks[i].x, t.blocks[i].y) {
			check = check || false
		} else {
			check = check || true
		}
	}

	if !check {
		return false
	}
	return true
}

//创建一个新的block
func (t *tankGame) drawNewBlock() {
	var coords = t.generateCoords()
	var X = coords[0]
	var Y = coords[1]
	var width = 40
	var height = 60

	for t.blockDistanceChecker(X, Y) {
		coords = t.generateCoords()
		X = coords[0]
		Y = coords[1]
	}
	t.blocks = append(t.blocks, block{X, Y, width, height})
}

//绘画所有的block
func (t *tankGame) drawBlocks() {
	ctx := t.ctx
	for i := 0; i < len(t.blocks); i++ {
		ctx.BeginPath()
		ctx.Rect(t.blocks[i].x, t.blocks[i].y, t.blocks[i].width, t.blocks[i].height)
		ctx.SetFillStyle("green")
		ctx.Fill()
		ctx.ClosePath()
	}
}

//block 移动函数
func (t *tankGame) moverFunc() {
	for i := 0; i < len(t.blocks); i++ {
		t.blocks[i].y = t.blocks[i].y + t.tank_speed

		if t.blocks[i].y > float64(t.canvas.Width()) {
			t.blocks = append(t.blocks[:i], t.blocks[i+1:]...)
		}
	}
}

//ball 移动函数 master移动函数
func (t *tankGame) moveBalls() {
	for i := 0; i < len(t.balls); i++ {
		t.balls[i].y = t.balls[i].y - t.ball_speed

		if t.balls[i].y < 0 {
			t.balls = append(t.balls[:i], t.balls[i+1:]...)
		}
	}

	for i := 0; i < len(t.monsters); i++ {
		t.monsters[i].y = t.monsters[i].y + t.monsters_speed

		if t.monsters[i].y > float64(t.canvas.Height()) {
			t.monsters = append(t.monsters[:i], t.monsters[i+1:]...)
		}
	}
}

// tank 和 block 是否碰撞
func (t *tankGame) tank_block_collision() {
	for i := 0; i < len(t.blocks); i++ {
		conflict_X := false
		conflict_Y := false

		if t.tank.x+float64(t.tank.width) > t.blocks[i].x && t.tank.x < t.blocks[i].x+40 {
			conflict_X = conflict_X || true
		}

		if t.tank.y > t.blocks[i].y && t.tank.y < t.blocks[i].y+60 {
			conflict_Y = conflict_Y || true
		}

		if conflict_X && conflict_Y {
			t.tank_block_collision_bool = false
			t.player_lives -= 1
			return
		}
	}
	t.tank_block_collision_bool = true
}

// 创建怪兽
func (t *tankGame) create_monster() {
	var coords = t.generateCoords()
	var X = coords[0]
	var Y = coords[1]

	t.monsters = append(t.monsters, block{X, Y, 25, 29})
}

// 绘画怪兽
func (t *tankGame) draw_monster(X, Y float64) {
	var scale = 0.8
	var h float64 = 9
	var a float64 = 5
	var ctx = t.ctx

	ctx.BeginPath()
	ctx.MoveTo(X, Y)
	ctx.LineTo(X-a*scale, Y+h*scale)
	ctx.LineTo(X+(a+4)*scale, Y+h*scale)
	ctx.LineTo(X+(a+3)*scale, Y+h*scale)

	ctx.MoveTo(X-(a+5)*scale, Y+h*scale)
	ctx.LineTo(X-(a)*scale, Y+(h+20)*scale)
	ctx.LineTo(X+(a+15)*scale, Y+(h+20)*scale)
	ctx.LineTo(X+(a+20)*scale, Y+(h)*scale)
	ctx.SetFillStyle("purple")

	ctx.Fill()
	ctx.ClosePath()
}

//绘画所得的怪物
func (t *tankGame) draw_monsters() {
	for i := 0; i < len(t.monsters); i++ {
		X := t.monsters[i].x
		Y := t.monsters[i].y

		t.draw_monster(X, Y)
	}
}

//块与块之间的碰撞
func (t *tankGame) collision_detector(first, second block) bool {
	var x1 = first.x
	var y1 = first.y
	var width1 = float64(first.width)
	var height1 = float64(first.height)
	var x2 = second.x
	var y2 = second.y
	var width2 = float64(second.width)
	var height2 = float64(second.height)

	if x2 > x1 && x2 < x1+width1 || x1 > x2 && x1 < x2+width2 {
		if y2 > y1 && y2 < y1+height1 || y1 > y2 && y1 < y2+height2 {
			return true
		}
	}

	return false
}

//ball碰撞master
func (t *tankGame) ball_monster_collision() {
	for i := 0; i < len(t.monsters); i++ {
		var monster = t.monsters[i]
		for j := 0; j < len(t.balls); j++ {
			var ball = t.balls[j]
			if t.collision_detector(monster, ball) {
				t.monsters = append(t.monsters[:i], t.monsters[i+1:]...)
				t.balls = append(t.balls[:i], t.balls[i+1:]...)
				t.score += 2
			}
		}
	}
}

//填充文字
func (t *tankGame) drawInfo() {
	ctx := t.ctx
	ctx.SetFont("bold 18px Gill Sans MT")
	ctx.SetFillStyle("blue")
	ctx.FillText("Lives: "+strconv.Itoa(t.player_lives), float64(t.canvas.Width())-45, 22)
	ctx.FillText("Sorce: "+strconv.Itoa(t.score), 13, 22)
}

// 绘画
func draw() {
	game.ctx.ClearRect(0, 0, game.canvas.Width(), game.canvas.Height())
	game.drawBorder()
	game.drawInfo()
	game.drawTank()
	game.drawBalls()
	game.drawBlocks()
	game.draw_monsters()
	game.moveBalls()
	game.tank_block_collision()
	game.ball_monster_collision()
	game.moverFunc()

	if game.space_pressed && len(game.balls) < 10 && timenow()-game.since_last_fire > 500 {
		game.drawNewBall(game.tank.x+15, game.tank.y-30)
	}

	if game.right_pressed && game.tank.x+float64(game.tank.width) < float64(game.canvas.Width()) {
		game.tank.x = game.tank.x + 1
	}

	if game.left_pressed && game.tank.x > 0 {
		game.tank.x = game.tank.x - 1
	}

	if len(game.blocks) < 3 {
		fmt.Println(game.blocks)
		game.drawNewBlock()
		fmt.Println(game.blocks)
	}

	if len(game.monsters) < 1 {
		game.create_monster()
	}

	if !game.tank_block_collision_bool || game.player_lives < 0 {
		web.Window.Alert("you lost.")
		web.Document.Location().Reload()
		return
	}

	web.Window.RequestAnimationFrame(js.NewCallback(func(args []js.Value) {
		draw()
	}))
}

func init() {
	rand.Seed(time.Now().Unix())

	web.Document.AddEventListener("keydown", func(args []js.Value) {
		if len(args) > 0 {
			keyCode := args[0].Get("keyCode").Int()
			switch keyCode {
			case 39:
				game.right_pressed = true
			case 37:
				game.left_pressed = true
			case 32:
				game.space_pressed = true
			}
		}
	})

	web.Document.AddEventListener("keyup", func(args []js.Value) {
		if len(args) > 0 {
			keyCode := args[0].Get("keyCode").Int()
			switch keyCode {
			case 39:
				game.right_pressed = false
			case 37:
				game.left_pressed = false
			case 32:
				game.space_pressed = false
			}
		}
	})

	game = newTankGame()
}

func main() {
	fmt.Println("hello tank")

	draw()

	select {}

	fmt.Println("code had exit")
}
