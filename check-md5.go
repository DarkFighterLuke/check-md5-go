package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	scriptDataPath = ".check_md5/.resume_data"
)

var mutex = &sync.Mutex{}

var startIndex = 0
var resumeFile *os.File
var resumeData = make([]string, 3)

func main() {
	if len(os.Args) != 4 {
		panic("Wrong parameters number!\n" +
			"These 3 parameters are needed:\n" +
			" - First checksum file\n" +
			" - Second checksum file\n" +
			" - File where to write negative results of the comparison\n")
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		sig := <-sigs
		fmt.Println("\nReceived " + sig.String() + ". Saving resume data...")
		writeResumeData(resumeData)
		os.Exit(0)
	}()

	os.MkdirAll(".check_md5", 0766)
	filepath1 := os.Args[1]
	filepath2 := os.Args[2]
	resultspath := os.Args[3]
	file1, err := os.Open(filepath1)
	if err != nil {
		panic("Could not open the first file.")
	}
	defer file1.Close()
	file2, err := os.Open(filepath2)
	if err != nil {
		panic("Could not open the second file.")
	}
	defer file2.Close()
	fileData1 := readFileContent(filepath1)
	fileData2 := readFileContent(filepath2)

	results, err := os.OpenFile(resultspath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic("Could not create results file.")
	}
	defer results.Close()

	if _, err := os.Stat(scriptDataPath); !os.IsNotExist(err) {
		resumeFile, err = os.OpenFile(scriptDataPath, os.O_WRONLY, 0666)
		if err != nil {
			panic("Could not open resume data file.")
		}
		defer resumeFile.Close()
		log.Println("Parsing resume data...")
		resumeData = readFileContent(scriptDataPath)
		if len(resumeData) != 4 {
			resumeData = make([]string, 3)
			log.Println("Corrupted resume data file. Deleting...")
			err = os.Remove(scriptDataPath)
			if err != nil {
				log.Println("Could not remove corrupted resume data file. Continuing will lead to unresumable data...")
			}
			resumeFile, err = os.Create(scriptDataPath)
			if err != nil {
				log.Println("Could not replace corrupted resume data file. Continuing will lead to unresumable data...")
			}
			generateResumeData(filepath1, filepath2)
		} else {
			oldIntegrity1 := resumeData[0]
			oldIntegrity2 := resumeData[1]
			log.Println("Calculating md5 of the files to compare...")
			integrity1 := md5.Sum([]byte(filepath1))
			integrity2 := md5.Sum([]byte(filepath2))
			if string(integrity1[:]) == oldIntegrity1 && string(integrity2[:]) == oldIntegrity2 {
				log.Println("Files to compare md5 matched! Starting from the last index before interruption.")
				startIndex, err = strconv.Atoi(resumeData[2])
				if err != nil {
					log.Println("Invalid index in resumed data. Starting from the first index.")
				}
			} else {
				log.Println("Files to compare md5 mismatch. Starting from the first index.")
			}
		}
	} else {
		log.Println("No old execution data found.")
		resumeFile, err = os.Create(scriptDataPath)
		if err != nil {
			log.Println("Could not replace corrupted resume data file. Continuing will lead to unresumable data...")
		}
		generateResumeData(filepath1, filepath2)
	}

	log.Println("Reordering record...")
	MergeSort(&fileData2)

	log.Println("\t--- Start of comparison ---\n")
	for i := startIndex; i < len(fileData1)-1; i++ {
		foundIndex := sort.SearchStrings(fileData2, fileData1[i])
		if fileData2[foundIndex] == fileData1[i] {
			log.Println(fileData1[i][:33] + "\t" + fileData1[i][34:] + "  OK" + "\t\t[" + strconv.Itoa(i+1) + "]")
		} else {
			log.Println(fileData1[i] + "\tMD5 mismatch or row not found!")
			_, err = results.WriteString(fileData1[i] + "\n")
			if err != nil {
				panic("Could not write on results file. Stopping...")
			}
		}
		resumeData[2] = strconv.Itoa(i)
		writeResumeData(resumeData)
	}
	log.Println("\n\t--- End of comparison ---")
	err = os.Remove(scriptDataPath)
	if err != nil {
		log.Println("Could not remove resume data file. Please delete it manually in " + scriptDataPath)
	}
}

func generateResumeData(filepath1, filepath2 string) {
	resumeFile, _ = os.Create(scriptDataPath)
	log.Println("Calculating md5 of the files to compare...")
	integrity1 := md5.Sum([]byte(filepath1))
	integrity2 := md5.Sum([]byte(filepath2))
	resumeData[0] = string(integrity1[:])
	resumeData[1] = string(integrity2[:])
}

func readFileContent(filename string) []string {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}
	fileData := strings.Split(string(fileBytes), "\n")
	return fileData
}

func writeResumeData(resumeData []string) {
	mutex.Lock()
	resumeFile, _ = os.Create(scriptDataPath)
	for _, r := range resumeData[:3] {
		_, err := resumeFile.WriteString(r + "\n")
		if err != nil {
			log.Println("Could not write on resume file. Resume data is not being saved!")
		}
	}
	resumeFile.Close()
	mutex.Unlock()
}
