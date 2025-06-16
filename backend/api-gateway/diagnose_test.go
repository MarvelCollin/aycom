package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("🔍 Diagnosing Go Test Issues")
	fmt.Println("=============================")

	// Test 1: Basic Go version
	fmt.Println("\n1. Checking Go version:")
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Error running 'go version': %v\n", err)
	} else {
		fmt.Printf("✅ %s", output)
	}

	// Test 2: Check if go test works with simple test
	fmt.Println("\n2. Testing basic go test functionality:")

	// Create a minimal test file
	testContent := `package simple_test

import "testing"

func TestSimple(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken!")
	}
}
`

	// Write test file
	cmd = exec.Command("cmd", "/c", "echo "+testContent+" > simple_test.go")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to create test file: %v\n", err)
		return
	}

	// Try to run the test with timeout
	cmd = exec.Command("go", "test", "-v", "-timeout", "5s", "simple_test.go")
	cmd.Dir = "."

	done := make(chan error, 1)
	go func() {
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("❌ go test failed: %v\n", err)
			fmt.Printf("Output: %s\n", output)
		} else {
			fmt.Printf("✅ go test worked:\n%s\n", output)
		}
		done <- err
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		fmt.Println("✅ Test completed")
	case <-time.After(10 * time.Second):
		fmt.Println("⏰ Test timed out - this indicates go test hangs")
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill process: %v", err)
		}
	}

	// Cleanup
	exec.Command("del", "simple_test.go").Run()

	fmt.Println("\n3. Diagnosis complete!")
}
