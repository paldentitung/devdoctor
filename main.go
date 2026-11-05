package main

import (
    "fmt"
    "os"

    "devdoctor/internal/registry"
    "devdoctor/internal/runner"
    "devdoctor/internal/system"
)

func main() {
    printSystemInfo()

    results := runner.Run(registry.Default())

    printResults(results)

    summary := runner.Summarize(results)
    fmt.Printf("Passed: %d  Warnings: %d  Failed: %d\n",
        summary.Passed, summary.Warnings, summary.Failed)

    if summary.Failed > 0 {
        os.Exit(1)
    }
}

func printResults(groups []runner.GroupResult) {
    for _, g := range groups {
        fmt.Println(g.Name)
        fmt.Println("------------------------------")

        for _, r := range g.Results {
            fmt.Printf("[%s] %s\n", r.Status, r.Check)
            fmt.Println(r.Message)

            if r.Suggestion != "" {
                fmt.Println("Suggestion:", r.Suggestion)
            }
            fmt.Println()
        }
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