# Noisy Data Pipeline (Go)

Implement a Go pipeline that processes a stream of integers mixed with noise. 
The pipeline must consist of several stages connected via channels and must stop cleanly when the context is canceled.

## Requirements

### 1. Noisy Generator
Create a generator that emits integers from `1` to `N`, but occasionally inserts the string `"noise"` into the output channel. Noise should appear roughly every 5–10 values.

### 2. Noise Filter
Implement a stage that removes all non-integer values. Only valid integers should pass through.

### 3. Transformer (Fan-Out)
Fan-out the filtered stream into **three parallel workers**.  
Each worker must multiply every incoming integer by `10`.

### 4. Aggregator (Fan-In)
Merge the results of the three workers back into a single output channel.

### 5. Context-Based Shutdown
The whole pipeline must stop when the provided `context.Context` is canceled.  
All goroutines must exit without leaks, and all channels must be closed correctly.

## Function to Implement

```go
func BuildNoisyPipeline(ctx context.Context, n int) <-chan int
```

## Constraints

- No use of time.Sleep in processing stages.
- Idiomatic Go: select, context cancellation, proper channel closing.
- No panics or data races.
- The pipeline must remain responsive to context cancellation.