package main

import (
	"bufio"
	"bytes"
	"strings"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/gin-gonic/gin"
)

func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}

func strToByte(s string)[]byte { return []byte(s)}

func getLogs(c *gin.Context) {
	url := c.Query("url")
	jsonStr := retrieveLogFromRepo(url)
	c.DataFromReader(http.StatusOK, int64(len(jsonStr)), gin.MIMEJSON, strings.NewReader(jsonStr), nil)
	fmt.Println("Response sent to client")
}

func main() {
		router := gin.Default()
		router.GET("/log", getLogs)
		router.Run(":8080")


		// url := "https://github.com/lppedd/idea-conventional-commit"
		// jsonStr := retrieveLogFromRepo((url))
}

func retrieveLogFromRepo(url string)string {
	var tmpDirLoc string

		fmt.Println("Downloading .git folder")
		cmd := exec.Command("git","clone","--no-checkout",url,"tmp")
		
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Download success...")
		fmt.Println("Changing working directory")
		
		// Change working diretory
		if wd, err := os.Getwd(); err != nil {
			log.Fatal(err)
			} else {
				tmpDirLoc = path.Join(wd, "tmp")
				err := os.Chdir(tmpDirLoc)
				if err != nil {
					log.Fatal(err)
				}
			}
			
		fmt.Println("Extracting log")
		
		var jsonLog string
		format := "format:commit_id%n%H%nparent_id%n%P%ncommit_name%n%cN%ncommit_email%n%cE%ncommit_date%n%ct%nsubject%n%s%nbody%n%b"
		cmd = exec.Command("git","log","--format="+format)
		if out, err := cmd.Output(); err != nil {
			log.Fatal(err)
		} else {
			jsonLog = parseLog(out)
		}
		
		// Removes the temporary folder created
		os.RemoveAll(tmpDirLoc)
		os.Chdir(path.Join(tmpDirLoc, ".."))
		fmt.Println("Removed temp folder")

		return jsonLog
}

func parseLog(logP []byte) string{
	keys := []string{"commit_id", "parent_id", "commit_name", "commit_email", "commit_date", "subject", "body", "commit_id"}
	byteKeys := Map(keys, strToByte)

	parsedLog := []map[string]string{}
	tmpData := make(map[string]string)
	buffer := []byte{}

	scanner := bufio.NewScanner(bytes.NewReader(logP))
	newLine := []byte("\n")
	currentIndex := -1
	// i:=0
	for scanner.Scan() {
        line := scanner.Bytes()

				// if i<50 {
				// 	fmt.Printf("%v\n",string(line))
				// 	i++
				// }

        if !bytes.Equal(line, byteKeys[currentIndex+1]) {
					buffer = append(buffer, line...)
					buffer = append(buffer, newLine...)
					continue
        }
				
				currentIndex += 1
				if (currentIndex == 0) {
					continue
				}
				
				tmpData[keys[currentIndex-1]] = strings.TrimRight(string(buffer), "\n")
				buffer = []byte{}
			
				if currentIndex == 7 {
					parsedLog = append(parsedLog, tmpData)
					tmpData = make(map[string]string)
					currentIndex = 0
				}
				
  }
	tmpData["body"] = string(buffer)
	parsedLog = append(parsedLog, tmpData)

  if err := scanner.Err(); err != nil {
    log.Fatal(err)
  }

	// fmt.Println("Parsed Content: ", parsedLog[:10])
	// return parsedLog


	jsonStr, err := json.Marshal(parsedLog)
	if err != nil {
		log.Fatal(err)
	}

	return string(jsonStr)
}