package crontab

import (
	"fmt"
	"time"

	"github.com/jasonlvhit/gocron"
)

func task() {
	fmt.Println("I am runnning task.", time.Now())
}
func superWang() {
	fmt.Println("I am runnning superWang.", time.Now())
}

func test(s *gocron.Scheduler, sc chan bool) {
	// time.Sleep(600 * time.Second)
	// s.Remove(task) //remove task
	time.Sleep(8 * time.Second)
	s.Clear()
	fmt.Println("All task removed")
	close(sc) // close the channel
}

func main() {
	scheduler := gocron.NewScheduler()
	// s.Every(1).Seconds().Do(task)
	scheduler.Every(1).Day().At("20:13").Do(task)
	scheduler.Every(4).Seconds().Do(superWang)

	sc := scheduler.Start() // keep the channel
	go test(scheduler, sc)  // wait
	<-sc                    // it will happens if the channel is closed
}
