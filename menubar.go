package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sqweek/dialog"
)

const MenuBarHeight = 20

type MenuItem struct {
	label  string
	action func()
}

type Menu struct {
	label string
	x     int  // define horizontal starting psotion of menu item
	open  bool // check if dropdown is visible
	items []MenuItem
}

type MenuBar struct {
	menus   []*Menu
	display *Display
}

func NewMenuBar(display *Display) *MenuBar {
	menuBar := &MenuBar{display: display}

	fileMenu := &Menu{
		label: "File",
		x:     5,
		items: []MenuItem{
			{
				label: "Open ROM",
				action: func() {
					// run file dialog in separate goroutine to avoid collsion with main thread
					go func() {
						path, err := dialog.File().
							Title("Load Chip8 ROM").
							Filter("Chip8 ROMs", "ch8", "rom", "bin").
							Filter("All Files", "*").
							Load()
						if err != nil {
							// user cancelled the dialog
							return
						}
						// create a new chip8 instance
						chip := NewChip(display.keypad)
						err = chip.LoadRom(path)
						if err != nil {
							log.Fatal(err)
						}
						display.chip = chip
						display.paused = false
					}()
				},
			},
		},
	}

	emuMenu := &Menu{
		label: "Emulation",
		x:     55,
		items: []MenuItem{
			{
				label: "Pause / Resume",
				action: func() {
					display.paused = !display.paused
				},
			},
		},
	}

	menuBar.menus = []*Menu{fileMenu, emuMenu}
	return menuBar
}

func (menuBar *MenuBar) Update() {
	mx, my := ebiten.CursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	if clicked {
		clickedOnMenu := false

		for _, menu := range menuBar.menus {
			if my >= 0 && my < MenuBarHeight && mx >= menu.x && mx < menu.x+80 {
				clickedOnMenu = true

				if menu.open {
					menu.open = false
				} else {
					for _, other := range menuBar.menus {
						other.open = false
					}
					menu.open = true
				}
			}
		}

		// check if menu item was clicked
		for _, menu := range menuBar.menus {
			if menu.open {
				for i, item := range menu.items {
					itemY := MenuBarHeight + i*20
					if my >= itemY && my < itemY+20 && mx >= menu.x && mx < menu.x+130 {
						item.action()
						menu.open = false
						clickedOnMenu = true
					}
				}
			}
		}

		// if click occured outside menu close opened menus
		if !clickedOnMenu {
			for _, m := range menuBar.menus {
				m.open = false
			}
		}
	}
}

func (menuBar *MenuBar) Draw(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 0, float32(ScreenWidth*menuBar.display.scale), MenuBarHeight, color.RGBA{40, 40, 40, 255}, false)

	mx, my := ebiten.CursorPosition()

	for _, menu := range menuBar.menus {
		hovered := my >= 0 && my < MenuBarHeight && mx >= menu.x && mx < menu.x+80
		if hovered || menu.open {
			vector.FillRect(screen, float32(menu.x-3), 0, 80, MenuBarHeight, color.RGBA{70, 70, 70, 255}, false)
		}
		ebitenutil.DebugPrintAt(screen, menu.label, menu.x, 5)

		if menu.open {
			vector.FillRect(screen, float32(menu.x), MenuBarHeight, 130, float32(len(menu.items)*20), color.RGBA{50, 50, 50, 255}, false)
			for i, item := range menu.items {
				itemY := MenuBarHeight + i*20
				if mx >= menu.x && mx < menu.x+130 && my >= itemY && my < itemY+20 {
					vector.FillRect(screen, float32(menu.x), float32(itemY), 130, 20, color.RGBA{80, 80, 180, 255}, false)
				}
				ebitenutil.DebugPrintAt(screen, item.label, menu.x+5, itemY+4)
			}
		}
	}
}
