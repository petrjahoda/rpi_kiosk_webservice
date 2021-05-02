package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServerIpAddress struct {
	ServerIpAddress string `json:"ServerIpAddress"`
	IpAddress       string `json:"IpAddress"`
	Mask            string `json:"Mask"`
	Gateway         string `json:"Gateway"`
	Dhcp            string `json:"Dhcp"`
	Connection      string `json:"Connection"`
}

type HomepageData struct {
	IpAddress       string
	Mask            string
	Gateway         string
	ServerIpAddress string
	Dhcp            string
	DhcpChecked     string
	Version         string
}

func GetNetworkData() (string, string, string, string, string, string) {
	if homepageLoaded {
		interfaceIpAddress := "not assigned"
		interfaceMask := "not assigned"
		interfaceGateway := "not assigned"
		interfaceDhcp := "no"
		backResult := "DATA:"
		active := "cable unplugged"
		data, _ := exec.Command("nmcli", "con", "show", "-active").Output()
		result := string(data)
		fmt.Println(result)
		for _, line := range strings.Split(strings.TrimSuffix(result, "\n"), "\n") {
			if strings.Contains(line, "ethernet") {
				active = "cable plugged"
			}
		}
		configDirectory := filepath.Join("/ro", "home", "pi", "config")
		configFileName := "config.json"
		configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
		readFile, _ := ioutil.ReadFile(configFullPath)
		ConfigFile := ServerIpAddress{}
		_ = json.Unmarshal(readFile, &ConfigFile)
		if !initiated {
			fmt.Println("DEBUG ", "NOT INITIATED")
			configFileUpdated := updateConfigFile(ConfigFile)
			if configFileUpdated {
				fmt.Println("DEBUG ", "CONFIG FILE UPDATED")
				readFile, _ := ioutil.ReadFile(configFullPath)
				ConfigFile := ServerIpAddress{}
				_ = json.Unmarshal(readFile, &ConfigFile)
				initiateConnection(ConfigFile)
				return "Updating...", "Updating...", "Updating...", "Updating...", active, ""
			} else {
				return "Plug in the cable, please...", "", "", "", active, ""
			}
		}
		if len(ConfigFile.Connection) == 0 {
			fmt.Println("DEBUG ", "CONFIG CONNECTION ZERO")
			configFileUpdated := updateConfigFile(ConfigFile)
			if configFileUpdated {
				return "Updating...", "Updating...", "Updating...", "Updating...", active, ""
			} else {
				return "Plug in the cable, please...", "", "", "", active, ""
			}
		} else {
			fmt.Println("DEBUG", "GOOD WAY")
			output, _ := exec.Command("nmcli", "con", "show", ConfigFile.Connection).Output()
			result := string(output)
			for _, line := range strings.Split(strings.TrimSuffix(result, "\n"), "\n") {
				if strings.Contains(line, "ipv4.method") {
					backResult += line + "|"
					interfaceDhcp = line[40:]
					if strings.Contains(interfaceDhcp, "auto") {
						interfaceDhcp = "yes"
					} else {
						interfaceDhcp = "no"
					}
				}
				if interfaceDhcp == "yes" {
					if strings.Contains(line, "IP4.ADDRESS") {
						backResult += line + "|"
						interfaceIpAddress = line[38:]
						interfaceIpAddress = interfaceIpAddress[:]
						splittedIpAddress := strings.Split(interfaceIpAddress, "/")
						maskNumber := splittedIpAddress[1]
						interfaceMask = CalculateMaskFrom(maskNumber)
					}
					if strings.Contains(line, "IP4.GATEWAY") {
						backResult += line + "|"
						interfaceGateway = line[40:]
						interfaceGateway = interfaceGateway[:]
					}
				} else {
					if strings.Contains(line, "ipv4.addresses") {
						backResult += line + "|"
						interfaceIpAddress = line[38:]
						interfaceIpAddress = interfaceIpAddress[:]
						splittedIpAddress := strings.Split(interfaceIpAddress, "/")
						maskNumber := splittedIpAddress[1]
						interfaceMask = CalculateMaskFrom(maskNumber)
					}
					if strings.Contains(line, "ipv4.gateway") {
						backResult += line + "|"
						interfaceGateway = line[40:]
						interfaceGateway = interfaceGateway[:]
					}
				}
			}
			if strings.Contains(interfaceGateway, "--") {
				interfaceGateway = "not assigned"
				interfaceIpAddress = "not assigned"
				interfaceMask = "not assigned"
			}
			if !strings.Contains(interfaceIpAddress, "assigned") {
				interfaceIpAddress = strings.ReplaceAll(interfaceIpAddress, " ", "")
			}
			if strings.Contains(interfaceIpAddress, "/") {
				interfaceIpAddress = strings.Split(interfaceIpAddress, "/")[0]
			}
			fmt.Println(backResult)
			return interfaceIpAddress, interfaceMask, interfaceGateway, interfaceDhcp, active, backResult
		}
	}
	return "", "", "", "", "", ""
}

func stopStream(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	streamSync.Lock()
	streamCanRun = false
	streamSync.Unlock()
	var responseData PasswordOutput
	responseData.Result = "ok"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func restartRpi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data PasswordInput
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		var responseData PasswordOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		return
	}
	if data.Password == "3600" {
		var responseData PasswordOutput
		responseData.Result = "ok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		result, err := exec.Command("reboot").Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(result)
		return
	}
	var responseData PasswordOutput
	responseData.Result = "nok"
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responseData)
}

func shutdownRpi(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var data PasswordInput
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		var responseData PasswordOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		return
	}
	if data.Password == "3600" {
		var responseData PasswordOutput
		responseData.Result = "ok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
		result, err := exec.Command("poweroff").Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(result)
		return
	}
	var responseData PasswordOutput
	responseData.Result = "nok"
	w.Header().Set("Content-Type", "application/json")

}

func indexPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	streamSync.Lock()
	streamCanRun = true
	streamSync.Unlock()
	_ = r.ParseForm()
	interfaceIpAddress, interfaceMask, interfaceGateway, dhcpEnabled, _, _ := GetNetworkData()
	interfaceServerIpAddress := LoadSettingsFromConfigFile()
	tmpl := template.Must(template.ParseFiles("html/index.html"))
	data := HomepageData{
		IpAddress:       interfaceIpAddress,
		Mask:            interfaceMask,
		Gateway:         interfaceGateway,
		Dhcp:            "no",
		ServerIpAddress: interfaceServerIpAddress,
		Version:         version,
	}
	if dhcpEnabled == "yes" {
		data.Dhcp = "yes"
	}
	homepageLoaded = true
	_ = tmpl.Execute(w, data)
}

func CalculateMaskFrom(maskNumber string) string {
	switch maskNumber {
	case "1":
		return "128.0.0.0"
	case "2":
		return "192.0.0.0"
	case "3":
		return "224.0.0.0"
	case "4":
		return "240.0.0.0"
	case "5":
		return "248.0.0.0"
	case "6":
		return "252.0.0.0"
	case "7":
		return "254.0.0.0"
	case "8":
		return "255.0.0.0"
	case "9":
		return "255.128.0.0"
	case "10":
		return "255.192.0.0"
	case "11":
		return "255.224.0.0"
	case "12":
		return "255.240.0.0"
	case "13":
		return "255.248.0.0"
	case "14":
		return "255.252.0.0"
	case "15":
		return "255.254.0.0"
	case "16":
		return "255.255.0.0"
	case "17":
		return "255.255.128.0"
	case "18":
		return "255.255.192.0"
	case "19":
		return "255.255.224.0"
	case "20":
		return "255.255.240.0"
	case "21":
		return "255.255.248.0"
	case "22":
		return "255.255.252.0"
	case "23":
		return "255.255.254.0"
	case "24":
		return "255.255.255.0"
	case "25":
		return "255.255.255.128"
	case "26":
		return "255.255.255.192"
	case "27":
		return "255.255.255.224"
	case "28":
		return "255.255.255.240"
	case "29":
		return "255.255.255.248"
	case "30":
		return "255.255.255.252"
	case "31":
		return "255.255.255.254"
	case "32":
		return "255.255.255.255"
	}
	return "-"
}
