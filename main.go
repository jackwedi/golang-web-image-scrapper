package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Job struct {
	Jobname string   `json:"jobname"`
	Urls    []string `json:"urls"`
}

type Metric struct {
	Urls  []string `json:"urls"`
	Time  string   `json:"time"`
	Files int      `json:"filesCount"`
}

// Regexp that captures all the images urls
var regExpImages = regexp.MustCompile(`(http(s?):)([/|.|\w|\s|-])*\.(?:jpg|gif|png|svg)|<img[^>]+\bsrc=["']([^"']+)["']`)
var regExpImagesNested = regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`)

var metrics map[string]Metric

/*
REST endpoint that responds the elapsed time for the job given and dowloads images into current storage
*/
func crawl(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var job Job
	err2 := json.Unmarshal(body, &job)
	if err2 != nil {
		fmt.Println("error:", err2)
	}
	if !createOutputDir(job.Jobname) {
		w.Write([]byte("JOB EXISTS ALREADY, CHANGE JOB NAME"))
		return
	}

	var wg sync.WaitGroup

	start := time.Now()
	newMetric := Metric{Urls: job.Urls}

	for i := 0; i < len(job.Urls); i++ {
		wg.Add(1)
		// Makes the process of downloading images from one URL concurrent
		// This way if one user list a big list of URLS it won't slow the other users
		go downloadImages(job.Jobname, job.Urls[i], &wg, &newMetric)
	}
	wg.Wait()

	elapsed := time.Since(start)
	newMetric.Time = fmt.Sprint(elapsed)
	metrics[job.Jobname] = newMetric
	fmt.Printf("Done job: %s \n", job.Jobname)
	w.Write([]byte("Done job: " + job.Jobname + " in " + fmt.Sprint(elapsed)))
}

/*
Download the images contained in the website URL into a jobname folder.
Updates the metrics.
*/
func downloadImages(jobname, websiteURL string, wg *sync.WaitGroup, metric *Metric) {
	defer wg.Done()
	result := crawlImages(websiteURL)
	metric.Files = 0
	for i := 0; i < len(result); i++ {
		stringUrl := string(result[i])

		// Here I am using the index i to name the file and forcing to conversion to .png
		// When downloading files from Unsplash I did not manage to get some of the extensions and a correct filename
		var imgURL string
		var name string
		if strings.Contains(stringUrl, "<img") {
			// Need to extract nested urls
			imgURL = string(regExpImagesNested.Find(result[i]))
			name = jobname + "/" + fmt.Sprint(i) + ".png"
		} else {
			// Already usable urls
			imgURL = stringUrl
			splittedURL := strings.Split(stringUrl, "/")
			name = jobname + "/" + splittedURL[len(splittedURL)-1]
		}

		// fmt.Printf("RAW %s\n", stringUrl)
		// fmt.Printf("CLEANED %s %s\n\n", name, imgURL)

		// Only dowloading the file if it is a name longer than a char
		if len(imgURL) > 0 {
			err := downloadFile(imgURL, name)
			if err != nil {
				log.Fatal(err)
			} else {
				metric.Files += 1
			}
		}
	}
	fmt.Printf("\n")
}

/*
Returns all the images links contained in the body of the given url
*/
func crawlImages(url string) [][]byte {
	fmt.Printf("Crawling %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	imagesArr := regExpImages.FindAll(body, -1)

	return imagesArr
}

/*
Download the file located at URL and saves it at the current job directory
*/
func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

/*
Creates the directory with the dirname given
If the directory alreday exists returns false and does not create anything
*/
func createOutputDir(dirname string) bool {
	c, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	var dirAlreadyExist = false
	for _, entry := range c {
		if entry.Name() == dirname {
			dirAlreadyExist = true
			fmt.Println(" ", entry.Name(), entry.IsDir())
		}
	}

	if dirAlreadyExist {
		return false
	}

	err2 := os.Mkdir(dirname, 0755)
	if err2 != nil {
		log.Fatal(err2)
	}
	return true
}

/*
REST get endpoint that respond the current metrics file
*/
func getMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(metrics)
}

func main() {
	metrics = make(map[string]Metric)

	http.HandleFunc("/", crawl)
	http.HandleFunc("/metrics", getMetrics)
	log.Fatal(http.ListenAndServe(":1337", nil))
}
