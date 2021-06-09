package crontab

import (
	"fmt"

	"github.com/jasonlvhit/gocron"
)

func task() {
	fmt.Println("I am running task.")
}

func Init() {
	// gocron.Every(1).Day().At("20:29").Do(task)
	gocron.Every(20).Seconds().Do(task)
	_, time := gocron.NextRun()
	fmt.Println(time)
	<-gocron.Start()
}
