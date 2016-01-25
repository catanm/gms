package main

import (
	"gms/log"
	"gms/models"
	"gms/utils"
	"labix.org/v2/mgo/bson"
	"runtime"
	"strconv"
	"time"
)

var shutdown = false
var threads = 3

func main() {
	log.Info("Glasgow Memories Server Processing Module")

	m.Connect()
	defer m.Close()

	threads = runtime.NumCPU() - 2
	if threads < 2 {
		threads = 2
	}
	log.Info("Starting processing with " + strconv.Itoa(threads) + " threads.")
	// schedule the execution
	RunProcessing()
}

// This should be the work dispatching thread, only dispatching work between 0 and 7
func RunProcessing() {
	// create a channel for the pool
	imageChannel := make(chan m.Image, 50)
	// create a work done channel
	done := make(chan int, threads)

	counter := 0

	for true {
		var notProcessedImages []m.Image

		if counter < 1 {
			counter = 4
			count, err := m.GetDB("Image").Find(bson.M{"processed": false}).Count()
			if err != nil {
				log.Error("Could not count the unprocessed images")
			} else {
				log.Info("Total unprocessed images left: " + strconv.Itoa(count))
			}
		} else {
			counter--
		}

		m.GetDB("Image").Find(bson.M{"processed": false}).Limit(50).All(&notProcessedImages)
		if len(notProcessedImages) == 0 {
			log.Info("No images found, sleeping for 1 hour")
			counter = 0
			time.Sleep(1 * time.Hour)
			continue
		}

		for _, image := range notProcessedImages {
			imageChannel <- image
		}

		for i := threads; i > 0; i-- {
			go ProcessImage(imageChannel, done)
		}

		for i := threads; i > 0; i-- {
			<-done
		}
	}
}

func ProcessImage(images chan m.Image, done chan int) {
	for len(images) > 0 && !shutdown {
		image := <-images
		utils.LinkImageToGps(image, nil)
		utils.ImagePostprocessing(image, nil)
	}
	done <- 1
}
