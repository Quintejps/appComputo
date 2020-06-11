package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	bolt "go.etcd.io/bbolt"
)

type WorkerInfo struct {
	Name  string
	Tags  string
	Token string
	Usage string
	Port  int
}

var activeUsers = make(map[string]string)
var activeWorkers = make(map[string]WorkerInfo)
var work = make(chan string)
var doneWorker string

func loginVerification(user string) bool {
	for _, value := range activeUsers {
		if user == value {
			return false
		}
	}
	return true
}

func Login(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	temp := loginVerification(user)

	if !temp {
		c.JSON(http.StatusOK, gin.H{
			"message": "You already in " + user,
		})
	} else {
		token := (xid.New()).String()
		activeUsers[token] = user

		c.JSON(http.StatusOK, gin.H{
			"message": "Hi " + user + " welcome to the DPIP System",
			"token":   token,
		})
	}

}

func Logout(c *gin.Context) {
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")

	for key, value := range activeUsers {
		if token == key {
			user := value
			delete(activeUsers, key)
			c.JSON(http.StatusOK, gin.H{
				"message": "Bye " + user + ", your token has been revoked",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Status(c *gin.Context) {

	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")

	for key, value := range activeUsers {
		if token == key {
			refreshDB()
			user := value
			c.JSON(http.StatusOK, gin.H{
				"message": "Hi " + user + ", the DPIP System is Up and Running",
				"time":    time.Now().UTC().String(),
			})
			for key, value := range activeWorkers {
				c.JSON(http.StatusOK, gin.H{
					key: value,
				})
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func StatusWorker(c *gin.Context) {
	name := c.Param("name")
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")

	for key := range activeUsers {
		if token == key {
			refreshDB()
			for key, value := range activeWorkers {
				if key == name {
					c.JSON(http.StatusOK, gin.H{
						key: value,
					})
				}
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Upload(c *gin.Context) {

	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	workLoadID := c.GetHeader("WorkLoad-id")
	fileNameID := c.GetHeader("FileName-id")

	refreshDB()

	for _, value := range activeWorkers {
		if token == value.Token {
			inputFile, err := os.Open("./cuda/" + fileNameID)
			if err != nil {
				fmt.Errorf("Couldn't open source file: %s", err)
			}
			outputFile, err := os.Create("./results/" + workLoadID + "/" + fileNameID)
			if err != nil {
				inputFile.Close()
				fmt.Errorf("Couldn't open dest file: %s", err)
			}
			defer outputFile.Close()
			_, err = io.Copy(outputFile, inputFile)
			inputFile.Close()
			if err != nil {
				fmt.Errorf("Writing to output file failed: %s", err)
			}
			// The copy was successful, so now delete the original file
			err = os.Remove("./cuda/" + fileNameID)
			if err != nil {
				fmt.Errorf("Failed removing original file: %s", err)
			}
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Download(c *gin.Context) {

	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	workLoadID := c.GetHeader("WorkLoad-id")
	fileNameID := c.GetHeader("FileName-id")

	refreshDB()

	for _, value := range activeWorkers {
		if token == value.Token {
			url := "http://localhost:8080/results/" + workLoadID + "/" + fileNameID

			response, e := http.Get(url)
			if e != nil {
				log.Fatal(e)
			}
			defer response.Body.Close()

			//open a file for writing
			file, err := os.Create("./cuda/" + fileNameID)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, response.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Success!")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func Filter(c *gin.Context) {
	refreshDB()
	file, err := c.FormFile("file")
	if err != nil {
		log.Fatal(err)
	}

	workLoad := c.GetHeader("workload-id")
	token := strings.Trim(strings.TrimLeft(c.GetHeader("authorization"), "Bearer"), " ")
	usage := 0
	counter := 0
	for key := range activeUsers {
		if token == key {
			for _, value := range activeWorkers {
				usage, _ = strconv.Atoi(strings.Trim(value.Usage, "%"))
				if usage >= 100 {
					counter = counter + 1
				}

			}

			if counter == len(activeWorkers) {
				c.JSON(http.StatusOK, gin.H{
					"message": "There are not available workers",
				})

				return
			}

			temp := strings.Split(workLoad, "&")
			dirName := temp[0]
			temp2 := strings.Split(temp[1], "=")
			imageFilter := temp2[1]

			dirNumFiles := countFilesDir(dirName)

			nameFile := strconv.Itoa(dirNumFiles+1) + filepath.Ext(file.Filename)
			createDir(dirName)
			err = c.SaveUploadedFile(file, "./results/"+dirName+"/"+nameFile)
			if err != nil {
				log.Fatal(err)
			}
			msg := dirName + " " + imageFilter + " " + nameFile
			work <- msg

			//c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))

			path := "http://localhost:8080/results/" + dirName + "/"

			time.Sleep(1 * time.Second)

			c.JSON(http.StatusOK, gin.H{
				"Workload ID": dirName,
				"Filter":      imageFilter,
				"Job ID":      dirNumFiles + 1,
				"Status":      "Finished in " + doneWorker,
				"Results":     path,
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Token " + token + " does not exist",
	})
}

func refreshDB() {
	var workerInfo WorkerInfo

	for key := range activeWorkers {
		delete(activeWorkers, key)
	}

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyBucket"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			err := json.Unmarshal(v, &workerInfo)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("key=%s, value=%+v\n", k, v)
			activeWorkers[string(k)] = workerInfo
		}

		return nil
	})
}

func countFilesDir(dirName string) int {
	files, _ := ioutil.ReadDir("./results/" + dirName)
	return len(files)
}

func ListenNameWorker(nameWorker chan string) {
	for {
		select {
		case name := <-nameWorker:
			doneWorker = name
			fmt.Println(name)
		}
	}
}

func createDir(nameDir string) {
	if _, err := os.Stat("./results/" + nameDir); os.IsNotExist(err) {
		err = os.Mkdir("./results/"+nameDir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func Start(workLoads chan string) {
	r := gin.Default()
	work = workLoads

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"user1": "pass1",
		"user2": "pass2",
		"user3": "pass3",
	}))

	r.Use(static.Serve("/results/", static.LocalFile("./results/", true)))

	authorized.GET("/login", Login)
	r.GET("/logout", Logout)
	r.GET("/status", Status)
	r.GET("/status/:name", StatusWorker)
	r.GET("/upload", Upload)
	r.GET("/download", Download)
	r.GET("/workloads/filter", Filter)

	r.Run() // listen and serve on 0.0.0.0:8080
}
