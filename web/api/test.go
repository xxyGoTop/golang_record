package api

import "fmt"

type TankShot interface {
	shot()
}

type TankDriver interface {
	driver()
}

type TankShell interface {
	shell()
}

type Tank struct {
	shot   TankShot
	driver TankDriver
	shell  TankShell
}

func newTank(shot TankShot, driver TankDriver, shell TankShell) *Tank {
	return &Tank{
		shot:   shot,
		driver: driver,
		shell:  shell,
	}
}

func (tank *Tank) doWork() {
	tank.shot.shot()
	tank.driver.driver()
	tank.shell.shell()
}

type People1Shot struct {
	TankShot
}

func (people *People1Shot) shot() {
	fmt.Println("...shot")
}

type People2Shot struct {
	TankShot
}

func (people *People2Shot) shot() {
	fmt.Println("<<<shot")
}
