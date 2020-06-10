package main

import (
	"fmt"
	"log"
	"os/exec"

	api "./api"

	controller "./controller"
	scheduler "./scheduler"
)

func imageProcessorCUDA() {
	comand := "nvcc"
	arg0 := "-I"
	arg1 := "D:/opencv/build/include"
	arg2 := "-L"
	arg3 := "D:/opencv/build/x64/vc14/lib"
	arg4 := "-l"
	arg5 := "D:/opencv/build/x64/vc14/lib/opencv_world3410"
	arg6 := "-o"
	arg7 := "cuda/opencv-test"
	arg8 := "cuda/main.cu"

	cmd := exec.Command(comand, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print(string(stdout))
}

func main() {

	log.Println("Welcome to the Distributed and Parallel Image Processing System")
	imageProcessorCUDA()
	workLoadScheduler := make(chan string)
	workLoadAPI := make(chan string)
	workLoadController := make(chan string)
	workLoad := make(chan string)
	nameWorker := make(chan string)
	_nameWorker := make(chan string)

	// Start Controller
	go controller.Start()

	go controller.ListenWorkloads(workLoadController, workLoad)

	// Start Scheduler
	go scheduler.Start(workLoadScheduler, nameWorker)

	//Start Api
	go api.Start(workLoadAPI)

	go api.ListenNameWorker(_nameWorker)

	for {
		select {
		case work := <-workLoad:
			workLoadScheduler <- work
		case work := <-workLoadAPI:
			workLoadController <- work
		case work := <-nameWorker:
			_nameWorker <- work
		}
	}
}
