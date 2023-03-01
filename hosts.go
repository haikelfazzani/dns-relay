package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func findHost(hostname string) string {
	cmd := exec.Command("bash", "-c", "grep -Fx '0.0.0.0       "+hostname+"' /etc/hosts")

	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	// 0.0.0.0	example.com
	fields := strings.Split(string(stdout), "       ")
	return fields[0]
}
