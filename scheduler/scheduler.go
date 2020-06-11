package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	pb "github.com/CodersSquad/dc-labs/challenges/third-partial/proto"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

//const (
//	address     = "localhost:50051"
//	defaultName = "world"
//)

var activeWorkers = make(map[string]WorkerInfo)
var ad string
var workerInfo WorkerInfo
var nameWorker = make(chan string)

type Job struct {
	Address string
	RPCName string
}

type WorkerInfo struct {
	Name  string
	Tags  string
	Token string
	Usage string
	Port  int
}

func schedule(job Job) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(job.Address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: job.RPCName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Scheduler: RPC respose from %s : %s", job.Address, r.GetMessage())
}

func refreshDB() {

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for key := range activeWorkers {
		delete(activeWorkers, key)
	}

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyBucket"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := json.Unmarshal(v, &workerInfo)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("key=%s, value=%+v\n", k, workerInfo)
			activeWorkers[string(k)] = workerInfo
		}

		return nil
	})
}

func distributeJobs(workLoadID string) {
	cont := 0
	min := 0
	var workerUsage WorkerInfo
	var workerName string

	refreshDB()

	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		log.Fatal(err)
	}

	for index, value := range activeWorkers {
		str := re.ReplaceAllString(value.Usage, "")
		i, err := strconv.Atoi(str)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(i)
		if cont == 0 {
			min = i
			workerName = index
			workerUsage = value
		}
		if i < min {
			min = i // found another smaller value, replace previous value in min
			workerName = index
			workerUsage = value
		}
		cont++
	}

	fmt.Println(workerName)
	fmt.Println(workerUsage)

	str := re.ReplaceAllString(workerUsage.Usage, "")
	fmt.Println(str)
	usage, err := strconv.Atoi(str)
	if err != nil {
		log.Fatal(err)
	}
	usage += 20
	usageStr := strconv.Itoa(usage)
	newUsage := "%" + usageStr

	workLoadID += " " + newUsage

	nameWorker <- workerName

	job := Job{"localhost:" + strconv.Itoa(workerUsage.Port), workLoadID}
	go schedule(job)
}

func Start(workLoad chan string, _nameWorker chan string) error {
	nameWorker = _nameWorker
	for {
		select {
		case workLoadID := <-workLoad:
			distributeJobs(workLoadID)
		}
		//schedule(job)
	}
}
