package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	_ "time"
)

const MPV_SOCKET = "/home/p/.mpvsocket"

type Output struct {
	Text  string `json:"text"`
	Class string `json:"class,omitempty"`
	Alt   string `json:"alt,omitempty"`
}

func (o *Output) Dump() {
	out, err := json.Marshal(o)
	if err != nil {
		log.Println(err.Error())
		return
	}

	fmt.Println(string(out))
}

func reader(c io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		r := bufio.NewScanner(c)
		for r.Scan() {
			d := json.NewDecoder(strings.NewReader(r.Text()))
			if d.More() {
				var out map[string]interface{}
				err := d.Decode(&out)
				if err != nil {
					if err == io.EOF {
						return
					}
					log.Println(err.Error())
					return
				}

				if val, ok := out["event"]; ok {
					if val == "property-change" {
						title := out["data"].(string)
						o := &Output{Text: title, Class: "mpv", Alt: title}
						o.Dump()
					}
				}
			}
		}
	}
}

func main() {
	c, err := net.Dial("unix", MPV_SOCKET)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	command := `{"command":["observe_property_string",2,"media-title"]}` + "\n"
	_, err = c.Write([]byte(command))
	if err != nil {
		log.Fatalln(err.Error())
	}

	go reader(c, &wg)

	wg.Wait()
}
