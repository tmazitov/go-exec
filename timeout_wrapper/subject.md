# Timeout Wrapper

## Description
You are given a function that performs a long-running operation (e.g., sends an HTTP request).  
Your task is to implement a wrapper that adds timeout support: if the operation does not finish within the specified time, the wrapper must stop waiting and return a timeout error.

## Requirements
- The wrapper must accept:
  - the original function to execute  
  - a timeout duration  
- The wrapper must return:
  - the original function’s result if it finishes in time  
  - an error indicating timeout if execution exceeded the limit  
- The function being wrapped cannot be modified.
- Use goroutines, channels, and standard library tools.

## Provided Code
```go
// longOperation simulates a long-running request.
// You cannot change this function.
func longOperation() (string, error) {
    time.Sleep(3 * time.Second)
    return "OK", nil
}
```