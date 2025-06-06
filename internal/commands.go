package internal

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

const MAX_OUTPUT_LENGTH = 1024 // Maximum length of command output before truncation

func RunCommandsParallel(commands []string) map[string]string {
    return RunCommandsParallelWithTimeout(commands, 15*time.Second)
}

func RunCommandsParallelWithTimeout(commands []string, timeout time.Duration) map[string]string {
    type result struct {
        cmd  string
        out  string
    }
    results := make(map[string]string)
    var wg sync.WaitGroup
    ch := make(chan result, len(commands))

    // Start a goroutine to close the channel when all workers are done
    go func() {
        wg.Wait()
        close(ch)
    }()

    for _, c := range commands {
        wg.Add(1)
        go func(cmd string) {
            defer wg.Done()
            ctx, cancel := context.WithTimeout(context.Background(), timeout)
            defer cancel()
            
            out, err := exec.CommandContext(ctx, "bash", "-c", cmd).CombinedOutput()
            output := string(out)
            if len(output) > MAX_OUTPUT_LENGTH {
                output = output[:MAX_OUTPUT_LENGTH] + "...[truncated]"
            }
            if err != nil {
                ch <- result{cmd, output + "\n[ERROR] " + err.Error()}
            } else {
                ch <- result{cmd, output}
            }
        }(c)
    }

    for r := range ch {
        results[r.cmd] = r.out
    }
    return results
}
