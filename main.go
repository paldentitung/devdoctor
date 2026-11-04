package main

import (
    "fmt"
    "devdoctor/checks"
)

func main() {
    system := checks.GetSystemInfo()

    fmt.Println("DevDoctor - System Information")
    fmt.Println("------------------------------")

    fmt.Println("Operating System:", system.OS)
    fmt.Println("Architecture:", system.Architecture)
    fmt.Println("CPU Cores:", system.CPUCores)
    fmt.Println("Go Version:", system.GoVersion)
	fmt.Println("Home Directory:", system.HomeDirectory)
	fmt.Println("Username:", system.Username)
	fmt.Println("Shell:", system.Shell)
}