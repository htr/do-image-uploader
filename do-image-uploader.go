package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/digitalocean/godo"
)

var cliArgs struct {
	WaitUntilAvailable bool   `kong:"help='waits until image is available'"`
	APIToken           string `kong:"help='DO api token',required,env='DO_API_TOKEN'"`
	ImageFile          string `kong:"help='path to image file',required,type='existingfile'"`
	Region             string `kong:"help='image region',default='nyc3'"`
	Name               string `kong:"help='image name',required"`
	HTTPPort           int    `kong:"help='http port',default='5379'"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	kong.Parse(&cliArgs)

	addr, err := getPublicAddress()
	if err != nil {
		log.Panic("unable to get public address:", err)
	}

	addr = fmt.Sprintf("%s:%d", addr, cliArgs.HTTPPort)
	httpPath := "/" + randomString(6) + path.Base(cliArgs.ImageFile)

	url := "http://" + addr + httpPath

	go func() {
		http.HandleFunc(httpPath, func(w http.ResponseWriter, r *http.Request) {
			inputFile, err := os.Open(cliArgs.ImageFile)
			if err != nil {
				log.Panic("unable to open image file:", err)
			}
			io.Copy(w, inputFile)
		})

		http.ListenAndServe(addr, nil)
	}()

	doClient := godo.NewFromToken(cliArgs.APIToken)

	customImageReq := &godo.CustomImageCreateRequest{
		Name:         cliArgs.Name,
		Region:       cliArgs.Region,
		Url:          url,
		Distribution: "Unknown",
	}

	ctx := context.Background()

	img, _, err := doClient.Images.Create(ctx, customImageReq)
	if err != nil {
		log.Fatalln("unable to create custom image", err)
	}

	imgID := img.ID

	fmt.Printf("created image %d\n", imgID)

	for {
		time.Sleep(2 * time.Second)
		img, _, err = doClient.Images.GetByID(ctx, imgID)
		if err != nil {
			log.Println("unable to get image info", err)
		}
		if img.Status != "new" {
			fmt.Println("image has been downloaded")
			break
		}
	}

	if cliArgs.WaitUntilAvailable {
		for {
			time.Sleep(5 * time.Second)
			img, _, err = doClient.Images.GetByID(ctx, imgID)

			if err != nil {
				log.Println("unable to get image info", err)
			}

			if img.Status == "available" {
				fmt.Println("image is available")
				break
			}
		}
	}
	fmt.Printf("image %d is %s\n", imgID, img.Status)
}

func getPublicAddress() (string, error) {
	resp, err := http.Get("http://ifconfig.me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	address := strings.TrimSpace(string(body))
	return address, nil
}

//// shamelessly stolen from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
