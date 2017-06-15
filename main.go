package main

import "fmt"
import "github.com/hink/go-blink1"

func main() {
	d, e := blink1.OpenNextDevice()

	if e != nil {
		fmt.Printf("err: %s\n", e.Error())
		return
	}

	defer d.Close()

	fmt.Printf("starting, (green) (red) (blue)\n")
	for {
		var in string

		if _, e := fmt.Scanln(&in); e != nil {
			fmt.Printf("err: %s", e.Error())
			return
		}

		if in == "exit" {
			fmt.Printf("exiting...\n")
			break
		}

		var s blink1.State
		switch in {
		case "red":
			s = blink1.State{Red: 255}
		case "blue":
			s = blink1.State{Blue: 255}
		case "green":
			s = blink1.State{Green: 255}
		}

		d.SetState(s)

	}

	fmt.Printf("fin.\n")
	d.SetState(blink1.State{})
}
