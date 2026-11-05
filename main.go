package main

import (
    "fmt"
    "os"

    "devdoctor/internal/checks"
    "devdoctor/internal/packagemanager"
    "devdoctor/internal/runtimes"
    "devdoctor/internal/system"
    "devdoctor/internal/tools"
)

func main() {
    printSystemInfo()

    diagnosticChecks := []checks.Check{
        runtimes.Node{},
        packagemanager.Npm{},
        tools.Git{},
    }

    hasFailure := false

    for _, c := range diagnosticChecks {
        result := c.Run()

        fmt.Printf("[%s] %s\n", result.Status, result.Check)
        fmt.Println(result.Message)

        if result.Suggestion != "" {
            fmt.Println("Suggestion:", result.Suggestion)
        }
        fmt.Println()

        if result.Status == checks.StatusFailed {
            hasFailure = true
        }
    }

    if hasFailure {
        os.Exit(1)
    }
}

func printSystemInfo() {
    info := system.GetInfo()

    fmt.Println("DevDoctor - System Information")
    fmt.Println("------------------------------")
    fmt.Println("Operating System:", info.OS)
    fmt.Println("Architecture:", info.Architecture)
    fmt.Println("CPU Cores:", info.CPUCores)
    fmt.Println("Go Version:", info.GoVersion)
    fmt.Println("Home Directory:", info.HomeDirectory)
    fmt.Println("Username:", info.Username)
    fmt.Println("Shell:", info.Shell)
    fmt.Println()
}