package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	directoryPath = "ML/MachineLearningAPI"
)

type Parameters struct {
	I        string  `json:"i"`
	J        string  `json:"j"`
	K        string  `json:"k"`
	Accuracy float64 `json:"accuracy"`
	Images   string  `json:"images"`
}

var learningRate = []float64{0.001, 0.01, 0.1}
var numofLayers = []int{1, 2, 4}
var numofSteps = []int{1000, 2000, 4000}
var accuracy float64
var learning_rate float64
var steps int
var layer int

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home!\n"))
}

func trainHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home!\n"))
	go executeTraining()
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home!\n"))
	go executeTesting()
}

func executeTesting() {
	db, err := sql.Open("mysql",
		"root:root@tcp(127.0.0.1:3306)/machine_learning")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.QueryRow("select learning_rate,layer,steps,accuracy from accuracy where accuracy = (SELECT MAX(accuracy) FROM accuracy)").Scan(&learning_rate, &layer, &steps, &accuracy)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Calling Test Function with highest accuracy (%v) parameters, Learning Rate = %v, Layer = %v, Steps = %v \n ", accuracy, learning_rate, layer, steps)
	executeExperiment(learning_rate, layer, steps, db)
}

func executeTraining() {

	fmt.Printf("Training Model \n")
	db, err := sql.Open("mysql",
		"root:root@tcp(127.0.0.1:3306)/machine_learning")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for _, i := range learningRate {
		for _, j := range numofLayers {
			for _, k := range numofSteps {
				executeExperiment(i, j, k, db)
			}
		}
	}
}

func executeExperiment(i float64, j int, k int, db *sql.DB) {
	cmd := exec.Command("python", "train.py", "--i", fmt.Sprintf("%.6f", i), "--j", strconv.Itoa(j), "--k", strconv.Itoa(k), "--images", "/home/ubuntu/TrainingImages/")
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}
	var p Parameters
	str := strings.Replace(string(out), "'", "\"", -1)
	out = []byte(str)
	err = json.Unmarshal(out, &p)
	if err != nil {
		fmt.Println("Error Unmarshalling Error, ", err)
	}
	query := fmt.Sprintf("INSERT INTO accuracy (learning_rate, layer, steps, accuracy) VALUES ( %s, %s, %s, %s )", fmt.Sprintf("%.6f", i), strconv.Itoa(j), strconv.Itoa(k), fmt.Sprintf("%.6f", p.Accuracy))
	insert, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	defer insert.Close()
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func checkifValidImage(r *http.Request) bool {
	file, _, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buff := make([]byte, 512)
	_, err = file.Read(buff)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filetype := http.DetectContentType(buff)

	fmt.Println(filetype)
	if filetype == "image/jpeg" || filetype == "image/jpg" || filetype == "image/png" {
		return true
	}
	return false
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "Upload files\n")

	ok := checkifValidImage(r)
	if !ok {
		w.Write([]byte("Upload Failed. Format Not Supported. Formats supported are jpeg, jpg and png.\n"))
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	handler.Filename = "TrainingImages/" + handler.Filename

	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(f, file)

}
func main() {
	currentDir, _ := os.Getwd()
	createDirIfNotExist(currentDir + "/" + "TrainingImages")
	r := mux.NewRouter()
	r.HandleFunc("/", defaultHandler)
	r.HandleFunc("/upload", uploadHandler)
	r.HandleFunc("/train", trainHandler)
	r.HandleFunc("/test", testHandler)
	go log.Fatal(http.ListenAndServe(":8000", r))

}
