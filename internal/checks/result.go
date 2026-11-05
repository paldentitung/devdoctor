package checks

type Status string

const (
    StatusPassed  Status = "passed"
    StatusWarning Status = "warning"
    StatusFailed  Status = "failed"
)

type Result struct {
    Check      string
    Status     Status
    Message    string
    Suggestion string
}

func Passed(name, msg string) Result {
    return Result{Check: name, Status: StatusPassed, Message: msg}
}
func Warning(name, msg, fix string) Result {
    return Result{Check: name, Status: StatusWarning, Message: msg, Suggestion: fix}
}
func Failed(name, msg, fix string) Result {
    return Result{Check: name, Status: StatusFailed, Message: msg, Suggestion: fix}
}