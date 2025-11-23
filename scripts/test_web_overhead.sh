#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TESTSERVER_BIN="./testserver"
XGOTOP_BIN="./xgotop"
OUTPUT_DIR="./web_overhead_test_results"
TEST_REQUESTS=10000
CURL_URL="http://localhost/books/test-book/page/100"

# Print colored message
print_msg() {
    local color=$1
    local msg=$2
    echo -e "${color}${msg}${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_msg "$YELLOW" "Checking prerequisites..."
    
    # Remove and create output directory
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
    
    print_msg "$GREEN" "Prerequisites OK"
}

run_test() {
    local readers_count=$1
    local processors_count=$2
    local web_mode=$3
    local metrics_file_prefix="readers_${readers_count}_processors_${processors_count}_web_${web_mode}"
    local output_prefix="${OUTPUT_DIR}/${metrics_file_prefix}"
    
    print_msg "$YELLOW" "\n=== Running test: $readers_count readers, $processors_count processors, web mode $web_mode ==="

    # Kill any existing processes
    sudo pkill -f "$TESTSERVER_BIN" 2>/dev/null || true
    sudo pkill -f "xgotop" 2>/dev/null || true
    sleep 2
    
    # Start testserver first
    print_msg "$GREEN" "Starting testserver..."
    $TESTSERVER_BIN 2>&1 &
    TESTSERVER_PID=$!
    
    # Give testserver time to start
    sleep 2
    
    # Check if testserver is running
    if ! kill -0 $TESTSERVER_PID 2>/dev/null; then
        print_msg "$RED" "Error: testserver failed to start"
        exit 1
    fi
    
    # Verify testserver is responding
    print_msg "$GREEN" "Verifying testserver is responding..."
    if ! curl -s -f -m 2 "$CURL_URL" > /dev/null 2>&1; then
        print_msg "$RED" "Error: testserver is not responding on $CURL_URL"
        sudo kill -INT $TESTSERVER_PID 2>/dev/null || true
        exit 1
    fi
    print_msg "$GREEN" "Testserver is ready!"
        
    if [ $web_mode -eq 1 ]; then
        print_msg "$GREEN" "Starting xgotop with web mode, $readers_count readers, $processors_count processors"
        sudo $XGOTOP_BIN -b $TESTSERVER_BIN -s -web -rw $readers_count -pw $processors_count -mfp "${metrics_file_prefix}" &
    else
        print_msg "$GREEN" "Starting xgotop with no web mode, $readers_count readers, $processors_count processors"
        sudo $XGOTOP_BIN -b $TESTSERVER_BIN -s -rw $readers_count -pw $processors_count -mfp "${metrics_file_prefix}" &
    fi
    
    XGOTOP_PID=$!
    
    # Give xgotop time to initialize
    sleep 2
    
    # Check if xgotop is still running
    if ! kill -0 $XGOTOP_PID 2>/dev/null; then
        print_msg "$RED" "Error: xgotop failed to start"
        sudo kill -INT $TESTSERVER_PID 2>/dev/null || true
        exit 1
    fi
    
    # Start curl loop
    print_msg "$GREEN" "Starting curl loop for $TEST_REQUESTS requests..."
    for i in $(seq 1 $TEST_REQUESTS); do
        curl -s "$CURL_URL" > /dev/null 2>&1
    done
    
    # Wait for all requests to complete
    sleep 3
    print_msg "$GREEN" "All requests completed!"
    
    # Kill testserver (we already have its PID)
    print_msg "$YELLOW" "Stopping testserver..."
    if [ ! -z "$TESTSERVER_PID" ] && kill -0 $TESTSERVER_PID 2>/dev/null; then
        sudo kill -INT $TESTSERVER_PID 2>/dev/null || true
        sleep 1
        # Force kill if still running
        if kill -0 $TESTSERVER_PID 2>/dev/null; then
            sudo kill -9 $TESTSERVER_PID 2>/dev/null || true
        fi
    fi
    
    # Give xgotop time to process remaining events
    sleep 2
    
    # Stop xgotop
    print_msg "$YELLOW" "Stopping xgotop..."
    sudo kill -INT $XGOTOP_PID 2>/dev/null || true
    
    # Wait for xgotop to write metrics
    wait $XGOTOP_PID 2>/dev/null || true
    
    # Find the metrics file (extract basename from output_prefix for the pattern)
    local metrics_file=$(find . -name "metrics_*${metrics_file_prefix}.json" -type f -mmin -1 | head -1)
    if [ -z "$metrics_file" ]; then
        print_msg "$RED" "Error: Could not find metrics file for $output_prefix"
        return 1
    fi
    
    # Move metrics file to output directory
    mv "$metrics_file" "${output_prefix}_metrics.json"
    print_msg "$GREEN" "Metrics saved to: ${output_prefix}_metrics.json"
    
    return 0
}

# Main test sequence
main() {
    print_msg "$GREEN" "\n🧪 XGOTOP Web Overhead Test Suite 🧪\n"
    
    check_prerequisites
    
    total_tests=$((5 * 5 * 2))
    current_test=0
    
    for readers in {1..5}; do
        for processors in {1..5}; do
            for web_mode in {0..1}; do
                current_test=$((current_test + 1))
                print_msg "$YELLOW" "\n>>> Progress: $current_test/$total_tests tests"
                
                run_test $readers $processors $web_mode
                
                # Small delay between tests
                sleep 2
            done
        done
    done
    
    print_msg "$GREEN" "\n✅ All tests completed!"
}

# Run main function
main
