#!/usr/bin/env python3
"""
Quick script to check metrics files and verify event counts are present.
"""

import sys
import json
import os

def check_metrics_file(filepath):
    """Check a single metrics file"""
    print(f"\n{'='*60}")
    print(f"File: {filepath}")
    print('='*60)
    
    if not os.path.exists(filepath):
        print(f"ERROR: File not found")
        return
    
    try:
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        # Check for event_counts
        if 'event_counts' in data:
            print("✓ event_counts field found!")
            event_type_names = {
                "0": "casgstatus",
                "1": "makeslice",
                "2": "makemap",
                "3": "newobject",
                "4": "newgoroutine",
                "5": "goexit"
            }
            
            print("\nEvent counts by type:")
            total = 0
            for event_id, count in sorted(data['event_counts'].items()):
                name = event_type_names.get(event_id, f"unknown({event_id})")
                print(f"  {name:<15}: {count:>10,}")
                total += count
            print(f"  {'TOTAL':<15}: {total:>10,}")
        else:
            print("✗ event_counts field NOT found")
            print("  This metrics file was generated with an older version of xgotop")
            print("  Rebuild xgotop and rerun the test to generate metrics with event counts")
        
        # Show other metrics summary
        print("\nOther metrics:")
        for key in ['rps', 'pps', 'ewp', 'lat']:
            if key in data and isinstance(data[key], list) and len(data[key]) > 0:
                values = [v for v in data[key] if v > 0]
                if values:
                    print(f"  {key}: max={max(values):.0f}, samples={len(values)}")
    
    except json.JSONDecodeError as e:
        print(f"ERROR: Invalid JSON: {e}")
    except Exception as e:
        print(f"ERROR: {e}")

def main():
    if len(sys.argv) < 2:
        print("Usage: python check_metrics.py <metrics_file1> [metrics_file2 ...]")
        print("\nExample:")
        print("  python check_metrics.py ./sampling_test_results/*.json")
        sys.exit(1)
    
    for filepath in sys.argv[1:]:
        check_metrics_file(filepath)

if __name__ == "__main__":
    main()
