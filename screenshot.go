package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"os/exec"
)

func screenshotPage(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	command := "sudo"
	args := []string{"-u", "pi", "maim", "image.png"}
	argumentDebug := ""
	for _, arg := range args {
		argumentDebug += arg + " "
	}
	result, err := exec.Command(command, args...).Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
	renderTemplate(w, "screenshot", &Page{})
}
