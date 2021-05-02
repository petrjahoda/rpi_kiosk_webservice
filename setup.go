package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ChangeInput struct {
	Password  string
	IpAddress string
	Mask      string
	Gateway   string
	Server    string
}

type ChangeOutput struct {
	Result string
}

func setupPage(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	streamSync.Lock()
	streamCanRun = false
	streamSync.Unlock()
	interfaceIpAddress, interfaceMask, interfaceGateway, dhcpEnabled, _, _ := GetNetworkData()
	interfaceServerIpAddress := LoadSettingsFromConfigFile()
	tmpl := template.Must(template.ParseFiles("html/setup.html"))
	data := HomepageData{
		IpAddress:       interfaceIpAddress,
		Mask:            interfaceMask,
		Gateway:         interfaceGateway,
		ServerIpAddress: interfaceServerIpAddress,
		Dhcp:            dhcpEnabled,
		DhcpChecked:     "",
		Version:         version,
	}
	if strings.Contains(dhcpEnabled, "yes") {
		data.DhcpChecked = "checked"
	}
	_ = tmpl.Execute(w, data)
}

func changeToStatic(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	configDirectory := filepath.Join("/ro", "home", "pi", "config")
	configFileName := "config.json"
	configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
	readFile, _ := ioutil.ReadFile(configFullPath)
	ConfigFile := ServerIpAddress{}
	_ = json.Unmarshal(readFile, &ConfigFile)
	if ConfigFile.Connection == "" {
		var responseData ChangeOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
	} else {
		var data ChangeInput
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			var responseData ChangeOutput
			responseData.Result = "nok"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(responseData)
			return
		}
		if data.Password == "3600" {
			pattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
			if pattern.MatchString(data.IpAddress) && pattern.MatchString(data.Gateway) {
				maskNumber := GetMaskNumberFrom(data.Mask)
				fmt.Println("CHANGING STATIC FOR: " + ConfigFile.Connection)
				output, err := exec.Command("nmcli", "con", "mod", ConfigFile.Connection, "ipv4.addresses", data.IpAddress+"/"+maskNumber, "ipv4.gateway", data.Gateway).Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(string(output))
				output, err = exec.Command("nmcli", "con", "mod", ConfigFile.Connection, "ipv4.method", "manual").Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(string(output))
				output, err = exec.Command("nmcli", "con", "down", ConfigFile.Connection).Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println("Terminal connection turned down: " + string(output))
				output, err = exec.Command("systemctl", "restart", "NetworkManager").Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println("Terminal connection turned up: " + string(output))
				output, err = exec.Command("mount", "-o", "remount,rw", "/ro").Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(output)
				configDirectory := filepath.Join("/ro", "home", "pi", "config")
				configFileName := "config.json"
				configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
				data := ServerIpAddress{
					ServerIpAddress: data.Server,
					IpAddress:       data.IpAddress,
					Mask:            data.Mask,
					Gateway:         data.Gateway,
					Dhcp:            "false",
					Connection:      ConfigFile.Connection,
				}
				file, _ := json.MarshalIndent(data, "", "  ")
				_ = ioutil.WriteFile(configFullPath, file, 0666)
				output, err = exec.Command("mount", "-o", "remount,ro", "/ro").Output()
				if err != nil {
					fmt.Println(err.Error())
				}
				fmt.Println(output)
			}
			var responseData ChangeOutput
			responseData.Result = "ok"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(responseData)
			return
		}
		var responseData PasswordOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
	}
}

func changeToDhcp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	configDirectory := filepath.Join("/ro", "home", "pi", "config")
	configFileName := "config.json"
	configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
	readFile, _ := ioutil.ReadFile(configFullPath)
	ConfigFile := ServerIpAddress{}
	_ = json.Unmarshal(readFile, &ConfigFile)
	if ConfigFile.Connection == "" {
		var responseData ChangeOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
	} else {
		var data ChangeInput
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			var responseData ChangeOutput
			responseData.Result = "nok"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(responseData)
			return
		}
		if data.Password == "3600" {
			fmt.Println("SAVING DHCP")
			output, err := exec.Command("nmcli", "con", "mod", ConfigFile.Connection, "ipv4.method", "auto").Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("SAVING DHCP RESULT: " + string(output))
			output, err = exec.Command("nmcli", "con", "down", ConfigFile.Connection).Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Terminal connection turned down: " + string(output))
			output, err = exec.Command("systemctl", "restart", "NetworkManager").Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("Terminal connection turned up: " + string(output))
			output, err = exec.Command("mount", "-o", "remount,rw", "/ro").Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("SAVING DHCP WRITE RESULT: " + string(output))
			configDirectory := filepath.Join("/ro", "home", "pi", "config")
			configFileName := "config.json"
			configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
			fmt.Println("FILEPATH: " + configFullPath)
			data := ServerIpAddress{
				ServerIpAddress: data.Server,
				Dhcp:            "true",
				Connection:      ConfigFile.Connection,
			}
			file, _ := json.MarshalIndent(data, "", "  ")
			_ = ioutil.WriteFile(configFullPath, file, 0666)
			output, err = exec.Command("mount", "-o", "remount,ro", "/ro").Output()
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Println("SAVING DHCP READ RESULT: " + string(output))
			var responseData ChangeOutput
			responseData.Result = "ok"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(responseData)
			return
		}
		var responseData ChangeOutput
		responseData.Result = "nok"
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(responseData)
	}

}

func GetMaskNumberFrom(maskNumber string) string {
	switch maskNumber {
	case "128.0.0.0":
		return "1"
	case "192.0.0.0":
		return "2"
	case "224.0.0.0":
		return "3"
	case "240.0.0.0":
		return "4"
	case "248.0.0.0":
		return "5"
	case "252.0.0.0":
		return "6"
	case "254.0.0.0":
		return "7"
	case "255.0.0.0":
		return "8"
	case "255.128.0.0":
		return "9"
	case "255.192.0.0":
		return "10"
	case "255.224.0.0":
		return "11"
	case "255.240.0.0":
		return "12"
	case "255.248.0.0":
		return "13"
	case "255.252.0.0":
		return "14"
	case "255.254.0.0":
		return "15"
	case "255.255.0.0":
		return "16"
	case "255.255.128.0":
		return "17"
	case "255.255.192.0":
		return "18"
	case "255.255.224.0":
		return "19"
	case "255.255.240.0":
		return "20"
	case "255.255.248.0":
		return "21"
	case "255.255.252.0":
		return "22"
	case "255.255.254.0":
		return "23"
	case "255.255.255.0":
		return "24"
	case "255.255.255.128":
		return "25"
	case "255.255.255.192":
		return "26"
	case "255.255.255.224":
		return "27"
	case "255.255.255.240":
		return "28"
	case "255.255.255.248":
		return "29"
	case "255.255.255.252":
		return "30"
	case "255.255.255.254":
		return "31"
	case "255.255.255.255":
		return "32"
	}
	return "0"
}

func CheckServerIpAddress(interfaceServerIpAddress string) bool {
	seconds := 2
	timeOut := time.Duration(seconds) * time.Second
	_, err := net.DialTimeout("tcp", interfaceServerIpAddress, timeOut)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
