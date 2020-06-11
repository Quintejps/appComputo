package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	pb "github.com/CodersSquad/dc-labs/challenges/third-partial/proto"
	"go.nanomsg.org/mangos"
	"google.golang.org/grpc"

	// register transports
	bolt "go.etcd.io/bbolt"
	"go.nanomsg.org/mangos/protocol/respondent"
	_ "go.nanomsg.org/mangos/transport/all"
)

var WorkLoad string
var workerUsage string = "%0"
var workerInfo WorkerInfo
var WorkLoadID, FileNameID string
var WorkLoadFilter string

type WorkerInfo struct {
	Name  string
	Tags  string
	Token string
	Usage string
	Port  int
}

var (
	defaultRPCPort = 50051
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

var (
	controllerAddress = ""
	workerName        = ""
	tags              = ""
	APIAddress        = ""
	storeToken        = ""
)

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		fmt.Println("- Run Clean Up - Delete Our Example File")
		db, err := bolt.Open("../my.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		db.Update(func(tx *bolt.Tx) error {
			tx.DeleteBucket([]byte("MyBucket"))
			return err
		})
		fmt.Println("- Good bye!")
		os.Exit(0)
	}()
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("RPC: Received: %v", in.GetName())
	WorkLoad = in.GetName()
	data := strings.Split(WorkLoad, " ")
	WorkLoadID = data[0]
	WorkLoadFilter = data[1]
	FileNameID = data[2]
	workerUsage = data[3]
	go startWorkLoad()
	return &pb.HelloReply{Message: "Workload: " + in.GetName() + "finished in worker " + workerInfo.Name}, nil
}

func imageProcessing() {
	comand := "cuda/opencv-test.exe"
	arg0 := "cuda/" + FileNameID
	arg1 := WorkLoadFilter
	cmd := exec.Command(comand, arg0, arg1)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println("Error consola")
		fmt.Print(string(stdout))
		fmt.Println(err.Error())
		return
	}
	fmt.Print(string(stdout))
}

func startWorkLoad() {
	url := "http://" + APIAddress + "/download"

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + workerInfo.Token

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Erro antes de CUDA 1")
		log.Println("Error on response.\n[ERRO] -", err)
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Add("WorkLoad-id", WorkLoadID)
	req.Header.Add("FileName-id", FileNameID)
	fmt.Println(FileNameID + ", " + WorkLoadID + ", " + WorkLoadFilter)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erro antes de CUDA 2")
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println(string([]byte(body)))

	// Comand promt CUDA
	imageProcessing()

	url = "http://" + APIAddress + "/upload"

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("1 Error depues de CUDA")
		log.Println("Error on response.\n[ERRO] -", err)
	}

	req.Header.Add("Authorization", bearer)
	req.Header.Add("WorkLoad-id", WorkLoadID)
	req.Header.Add("FileName-id", FileNameID)

	// Send req using http Client
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Println("2 Error depues de CUDA")
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ = ioutil.ReadAll(resp.Body)
	log.Println(string([]byte(body)))

	var updateUsage int

	ar, e := regexp.Compile(`[^\w]`)
	if e != nil {
		log.Println("3 Error depues de CUDA")
		log.Fatal(err)
	}

	str := ar.ReplaceAllString(workerUsage, "")

	updateUsage, err = strconv.Atoi(str)
	if err != nil {
		log.Println("4 Error depues de CUDA")
		log.Fatal(err)
	}

	updateUsage -= 20
	update := strconv.Itoa(updateUsage)
	workerUsage = "%" + update
}

func init() {
	flag.StringVar(&controllerAddress, "controller", "tcp://localhost:40899", "Controller address")
	flag.StringVar(&workerName, "node-name", "hard-worker", "Worker Name")
	flag.StringVar(&tags, "tags", "gpu,superCPU,largeMemory", "Comma-separated worker tags")
	flag.StringVar(&APIAddress, "image-store-endpoint", "localhost:8080", "API's address")
	flag.StringVar(&storeToken, "image-store-token", "bq0fmhdimfajb76beri0", "Image store token")
}

// joinCluster is meant to join the controller message-passing server
func joinCluster(port int) {
	var sock mangos.Socket
	var err error
	var msgEncoded []byte

	if sock, err = respondent.NewSocket(); err != nil {
		die("can't get new respondent socket: %s", err.Error())
	}
	if err = sock.Dial(controllerAddress); err != nil {
		die("can't dial on respondent socket: %s", err.Error())
	}
	for {
		if msgEncoded, err = sock.Recv(); err != nil {
			die("Cannot recv: %s", err.Error())
		}
		fmt.Printf("CLIENT(%s): RECEIVED \"%s\" SURVEY REQUEST\n",
			workerName, string(msgEncoded))

		workerInfo.Name = workerName
		workerInfo.Tags = tags
		workerInfo.Token = storeToken
		workerInfo.Usage = workerUsage
		workerInfo.Port = port

		msgEncoded, err = json.Marshal(workerInfo)
		if err != nil {
			return
		}

		fmt.Printf("CLIENT(%s): SENDING DATE SURVEY RESPONSE\n", workerInfo.Name)
		if err = sock.Send(msgEncoded); err != nil {
			die("Cannot send: %s", err.Error())
		}
	}
}

func getAvailablePort() int {
	port := defaultRPCPort
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
		if err != nil {
			port = port + 1
			continue
		}
		ln.Close()
		break
	}
	return port
}

func main() {
	flag.Parse()
	setupCloseHandler()

	rpcPort := getAvailablePort()

	// Subscribe to Controller
	go joinCluster(rpcPort)

	// Setup Worker RPC Server
	log.Printf("Starting RPC Service on localhost:%v", rpcPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", rpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
