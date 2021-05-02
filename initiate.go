package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func initiateConnection(ConfigFile ServerIpAddress) {
	if ConfigFile.Dhcp == "true" {
		fmt.Println("Setting terminal to DHCP")
		output, err := exec.Command("nmcli", "con", "mod", ConfigFile.Connection, "ipv4.method", "auto").Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Terminal set to DHCP: " + string(output))
	} else {
		fmt.Println("Setting terminal to static")
		pattern := regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`)
		if pattern.MatchString(ConfigFile.IpAddress) && pattern.MatchString(ConfigFile.Gateway) {
			maskNumber := GetMaskNumberFrom(ConfigFile.Mask)
			output, err := exec.Command("nmcli", "con", "mod", ConfigFile.Connection, "ipv4.addresses", ConfigFile.IpAddress+"/"+maskNumber, "ipv4.gateway", ConfigFile.Gateway).Output()
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
		}
	}
	initiated = true
}

func updateConfigFile(ConfigFile ServerIpAddress) bool {
	output, err := exec.Command("nmcli", "con", "show").Output()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	result := string(output)
	if len(strings.Split(strings.TrimSuffix(result, "\n"), "\n")) > 2 {
		fmt.Println("We have multiple available connections")
		for index, line := range strings.Split(strings.TrimSuffix(result, "\n"), "\n") {
			if index > 0 {
				splitted := strings.Split(line, " ")
				if splitted[8] != "--" && splitted[6] == "ethernet" {
					fmt.Println("Active connection found, setting up connection")
					ConfigFile.Connection = splitted[0] + " " + splitted[1] + " " + splitted[2]
					output, err = exec.Command("mount", "-o", "remount,rw", "/ro").Output()
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println(string(output))
					configDirectory := filepath.Join("/ro", "home", "pi", "config")
					configFileName := "config.json"
					configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
					data := ServerIpAddress{
						ServerIpAddress: ConfigFile.ServerIpAddress,
						IpAddress:       ConfigFile.IpAddress,
						Mask:            ConfigFile.Mask,
						Gateway:         ConfigFile.Gateway,
						Dhcp:            ConfigFile.Dhcp,
						Connection:      ConfigFile.Connection,
					}
					file, _ := json.MarshalIndent(data, "", "  ")
					_ = ioutil.WriteFile(configFullPath, file, 0666)
					output, err = exec.Command("mount", "-o", "remount,ro", "/ro").Output()
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println("Config file updated: + " + string(output))
					return true
				}
			}
		}
		output, err = exec.Command("mount", "-o", "remount,rw", "/ro").Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(string(output))
		configDirectory := filepath.Join("/ro", "home", "pi", "config")
		configFileName := "config.json"
		configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
		data := ServerIpAddress{
			ServerIpAddress: ConfigFile.ServerIpAddress,
			IpAddress:       ConfigFile.IpAddress,
			Mask:            ConfigFile.Mask,
			Gateway:         ConfigFile.Gateway,
			Dhcp:            ConfigFile.Dhcp,
			Connection:      "",
		}
		file, _ := json.MarshalIndent(data, "", "  ")
		_ = ioutil.WriteFile(configFullPath, file, 0666)
		output, err = exec.Command("mount", "-o", "remount,ro", "/ro").Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("Config file updated: + " + string(output))
		return false
	} else {
		fmt.Println("We have one available connection")
		for index, line := range strings.Split(strings.TrimSuffix(result, "\n"), "\n") {
			if index == 1 {
				splitted := strings.Split(line, " ")
				if splitted[6] == "ethernet" {
					fmt.Println("Setting up connection")
					ConfigFile.Connection = splitted[0] + " " + splitted[1] + " " + splitted[2]
					output, err = exec.Command("mount", "-o", "remount,rw", "/ro").Output()
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println(output)
					configDirectory := filepath.Join("/ro", "home", "pi", "config")
					configFileName := "config.json"
					configFullPath := strings.Join([]string{configDirectory, configFileName}, "/")
					data := ServerIpAddress{
						ServerIpAddress: ConfigFile.ServerIpAddress,
						IpAddress:       ConfigFile.IpAddress,
						Mask:            ConfigFile.Mask,
						Gateway:         ConfigFile.Gateway,
						Dhcp:            ConfigFile.Dhcp,
						Connection:      ConfigFile.Connection,
					}
					file, _ := json.MarshalIndent(data, "", "  ")
					_ = ioutil.WriteFile(configFullPath, file, 0666)
					output, err = exec.Command("mount", "-o", "remount,ro", "/ro").Output()
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println("Config file updated: + " + string(output))
					return true
				}
			}
		}
		return false
	}
}
