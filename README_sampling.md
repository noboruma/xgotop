# xgotop Sampling Feature

## Overview

The sampling feature in xgotop allows you to reduce uprobe overhead by randomly sampling events rather than capturing every occurrence. This is particularly useful for high-frequency events that could cause performance issues or backpressure in the event queue.

## Implementation Details

### eBPF Side
- Added a `sampling_rates` BPF map to store sampling rates for each event type
- Added `CHECK_SAMPLING` macro that uses `bpf_get_prandom_u32()` for probabilistic sampling
- Each uprobe checks its sampling rate at the beginning and exits early if the event should be skipped

### User Space Side
- Added `--sample` flag to specify sampling rates
- Parses sampling configuration and updates the BPF map before attaching probes
- Supports per-event-type sampling rates

## Usage

### Basic Usage

```bash
# Sample all events at 10%
sudo ./xgotop -b /path/to/binary --sample 'newgoroutine:0.1,makemap:0.1,makeslice:0.1,newobject:0.1'

# Sample only specific events
sudo ./xgotop -b /path/to/binary --sample 'makemap:0.01,newgoroutine:0.5'

# Mixed sampling rates
sudo ./xgotop -b /path/to/binary --sample 'newgoroutine:0.8,makemap:0.2,makeslice:0.1,newobject:0.5'
```

### Sampling Rate Format

- Format: `event_name:rate,event_name:rate,...`
- Rate: Decimal between 0.0 and 1.0 (0% to 100%)
- Event names:
  - `casgstatus`: Goroutine status changes
  - `makeslice`: Slice allocations
  - `makemap`: Map allocations
  - `newobject`: Object allocations
  - `newgoroutine`: Goroutine creations
  - `goexit`: Goroutine exits

## Testing

### 1. Build Test Server

```bash
go build -o testserver ./cmd/testserver
```

### 2. Run Manual Tests

```bash
# First, start the testserver (needs sudo for port 80)
sudo ./testserver &

# In another terminal, start xgotop with baseline (no sampling)
sudo ./xgotop -b ./testserver -s -mfs "_baseline"

# Or with 10% sampling
sudo ./xgotop -b ./testserver -s -mfs "_10pct" --sample 'newgoroutine:0.1,makemap:0.1,makeslice:0.1,newobject:0.1'

# In a third terminal, generate load
while true; do curl -s http://localhost/books/test-book-31/page/62 > /dev/null; done
```

### 3. Automated Testing

```bash
# Run comprehensive test suite
./scripts/test_sampling.sh

# The test script will:
# - Start xgotop tracing testserver
# - Run a curl loop for 60 seconds per test
# - Collect metrics for different sampling configurations
# - Validate the results

# Validate results manually
python3 ./scripts/validate_sampling.py \
  ./sampling_test_results/baseline_metrics.json \
  ./sampling_test_results/uniform_10pct_metrics.json \
  "newgoroutine:0.1,makemap:0.1,makeslice:0.1,newobject:0.1"
```

## Performance Benefits

Sampling provides several benefits:

1. **Reduced Overhead**: Fewer uprobe executions mean less CPU overhead
2. **Less Backpressure**: Fewer events in the ringbuffer reduce the chance of drops
3. **Statistical Analysis**: For many use cases, a sample is sufficient for analysis
4. **Configurable**: Different events can have different sampling rates based on their frequency

## Example Scenarios

### High-Frequency Event Sampling
For applications with very high allocation rates:
```bash
--sample 'makeslice:0.01,makemap:0.01,newobject:0.01'
```

### Focus on Goroutines
When primarily interested in goroutine behavior:
```bash
--sample 'newgoroutine:1.0,goexit:1.0,makemap:0.1,makeslice:0.1,newobject:0.1'
```

### Production Monitoring
Low-overhead production monitoring:
```bash
--sample 'newgoroutine:0.05,makemap:0.01,makeslice:0.01,newobject:0.01'
```

## Validation

The validation script checks that actual sampling rates match expected rates within a tolerance:

```bash
python3 scripts/validate_sampling.py baseline.json sampled.json "event:rate,..." --tolerance 5
```

Expected output:
```
================================================================================
SAMPLING VALIDATION REPORT
================================================================================
Event Type      Baseline   Sampled    Expected %   Actual %     Error %    Status    
--------------------------------------------------------------------------------
makeslice       10000      1012       10.0         10.1         1.20       ✓ PASS    
makemap         10000      502        5.0          5.0          0.40       ✓ PASS    
newobject       10000      2489       25.0         24.9         0.44       ✓ PASS    
newgoroutine    100        78         80.0         78.0         2.50       ✓ PASS    
--------------------------------------------------------------------------------

Average Error: 1.14%
Pass Rate: 100.0%

✓ All sampling rates are within tolerance!
```

## Notes

- Sampling is probabilistic, so exact rates will vary slightly
- Very low sampling rates (< 1%) may show higher variance
- The overhead reduction is roughly proportional to (1 - sampling_rate)
- Events are sampled independently; there's no correlation between events
