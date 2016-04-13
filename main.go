package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/tarm/serial"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	go ReadFromUsb()
	fmt.Println("Press Ctrl+C to Exit")
	for {
		select {
		case <-quit:
			return
		}
	}
}

func ReadFromUsb() {

	usbPort := os.Getenv("USB_PORT")
	fmt.Println("Waiting For USB PORT : " + usbPort)
	c := &serial.Config{Name: usbPort, Baud: 38400}
	s, err := serial.OpenPort(c)
	panicIfErr(err)
	for {
		buffer := make([]byte, 256)
		buf := bytes.NewBuffer(buffer)
		count, err := io.CopyN(buf, s, 256)
		panicIfErr(err)
		panicIfErr(err)
		content := string(buf.Bytes())
		buf.Reset()
		fmt.Printf("Read %d bytes\n", count)
		fmt.Println("Received : " + content)
		b, _ := json.Marshal(&Message{content})
		SendData(string(b))
		buf.Reset()
	}
}

type Message struct {
	Msg string
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func SendData(data string) {
	url := os.Getenv("APP_URL")
	waterFlowApi := fmt.Sprintf("%s/data", url)
	req, _ := http.NewRequest("POST", waterFlowApi, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
