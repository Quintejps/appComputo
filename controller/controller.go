package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"go.nanomsg.org/mangos"

	// register transports
	"go.nanomsg.org/mangos/protocol/surveyor"
	_ "go.nanomsg.org/mangos/transport/all"
)

var controllerAddress = "tcp://localhost:40899"
var work = make(chan string)

type WorkerInfo struct {
	Name  string
	Tags  string
	Token string
	Usage string
	Port  int
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func date() string {
	return time.Now().Format(time.ANSIC)
}

func Start() {
	var workerInfo WorkerInfo
	var sock mangos.Socket
	var err error
	var msgEncoded []byte

	if sock, err = surveyor.NewSocket(); err != nil {
		die("can't get new surveyor socket: %s", err)
	}
	if err = sock.Listen(controllerAddress); err != nil {
		die("can't listen on surveyor socket: %s", err.Error())
	}
	err = sock.SetOption(mangos.OptionSurveyTime, time.Second/2)
	if err != nil {
		die("SetOption(): %s", err.Error())
	}
	for {
		time.Sleep(time.Second)
		fmt.Println("SERVER: SENDING DATE SURVEY REQUEST")
		if err = sock.Send([]byte("DATE")); err != nil {
			die("Failed sending survey: %s", err.Error())
		}
		for {
			if msgEncoded, err = sock.Recv(); err != nil {
				break
			}
			err = json.Unmarshal(msgEncoded, &workerInfo)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("SERVER: RECEIVED \"%+v\" SURVEY RESPONSE\n", workerInfo)
			msgEncoded, err = json.Marshal(workerInfo)
			if err != nil {
				return
			}
			go updateDB(msgEncoded, workerInfo.Name)
		}
		fmt.Println("SERVER: SURVEY OVER")
	}
}

func updateDB(msgEncoded []byte, workerName string) {
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		err := b.Put([]byte(workerName), msgEncoded)
		a := b.Get([]byte("answer"))
		fmt.Println(string(a))
		return err
	})
}

func startWorkload(_work string) {
	temp := strings.Split(_work, " ")

	dirName := temp[0]
	imageFilter := temp[1]
	fileName := temp[2]

	msg := dirName + " " + imageFilter + " " + fileName

	work <- msg

}

func ListenWorkloads(workLoad chan string, workLoadScheduler chan string) {
	work = workLoadScheduler
	for {
		select {
		case workID := <-workLoad:
			startWorkload(workID)
		}
	}
}
