package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/kardianos/service"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const version = "2021.2.2.2"
const programName = "Rpi kiosk webservice"
const programDescription = "Display some web page for rpi based terminals"

var initiated = false
var homepageLoaded = false

type Page struct {
	Title string
	Body  []byte
}

var streamCanRun = false
var streamSync sync.RWMutex

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	serviceConfig := &service.Config{
		Name:        programName,
		DisplayName: programName,
		Description: programDescription,
	}
	prg := &program{}
	s, _ := service.New(prg, serviceConfig)
	_ = s.Run()
}

func (p *program) run() {
	router := httprouter.New()
	networkDataStreamer := sse.New()
	router.GET("/image.png", image)
	router.GET("/", indexPage)
	router.GET("/screenshot", screenshotPage)
	router.GET("/setup", setupPage)
	router.POST("/password", checkPassword)
	router.POST("/restart", restartRpi)
	router.POST("/check_cable", checkCable)
	router.POST("/stop_stream", stopStream)
	router.POST("/shutdown", shutdownRpi)
	router.POST("/dhcp", changeToDhcp)
	router.POST("/static", changeToStatic)
	router.ServeFiles("/font/*filepath", http.Dir("font"))
	router.ServeFiles("/html/*filepath", http.Dir("html"))
	router.ServeFiles("/css/*filepath", http.Dir("css"))
	router.ServeFiles("/js/*filepath", http.Dir("js"))
	router.Handler("GET", "/networkdata", networkDataStreamer)
	go StreamNetworkData(networkDataStreamer)
	_ = http.ListenAndServe(":9999", router)
}

func checkCable(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, _, _, _, active, _ := GetNetworkData()
	var responseData ChangeOutput
	responseData.Result = active
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func LoadSettingsFromConfigFile() string {
	configDirectory := filepath.Join(".", "config")
	configFileName := "config.json"
	configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
	readFile, _ := ioutil.ReadFile(configFullPath)
	ConfigFile := ServerIpAddress{}
	_ = json.Unmarshal(readFile, &ConfigFile)
	ServerIpAddress := ConfigFile.ServerIpAddress
	return ServerIpAddress
}

func StreamNetworkData(streamer *sse.Streamer) {
	for {
		if streamCanRun {
			fmt.Println("streaming data")
			activeColor := "red"
			serverActiveColor := "red"
			interfaceIpAddress, interfaceMask, interfaceGateway, dhcpEnabled, active, result := GetNetworkData()
			interfaceServerIpAddress := LoadSettingsFromConfigFile()
			serverAccessible := CheckServerIpAddress(interfaceServerIpAddress)
			if dhcpEnabled == "yes" {
				dhcpEnabled = "yes"
			} else {
				dhcpEnabled = "no"
			}
			serverActive := "server not accessible"
			if serverAccessible {
				serverActive = "server accessible"
				serverActiveColor = "green"
			}
			if active == "cable plugged" {
				activeColor = "green"
			}
			streamer.SendString("", "networkdata", interfaceIpAddress+";"+interfaceMask+";"+interfaceGateway+";"+dhcpEnabled+";"+interfaceServerIpAddress+";"+result+";"+active+";"+serverActive+";"+activeColor+";"+serverActiveColor)
		}
		time.Sleep(5 * time.Second)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles("html/" + tmpl + ".html")
	_ = t.Execute(w, p)
}
func image(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	http.ServeFile(writer, request, "image.png")
}
