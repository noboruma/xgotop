# Neobrutalistic Metrics Plotter

A Python script for generating bold, geometric visualizations of performance metrics from JSON files.

## Features

- **Neobrutalistic Design**: Bold colors, thick borders, and geometric patterns
- **Multi-file Support**: Compare metrics across multiple JSON files
- **Automatic Layout**: 2x2 grid layout for performance metrics:
  - RPS: Reads Per Second
  - PPS: Processed Per Second
  - EWP: Events Waiting to be Processed
  - LAT: Latency (ms)
- **High-quality Output**: 300 DPI PNG images suitable for presentations

## Installation

Make sure you have activated the virtual environment and installed the required packages:

```bash
cd /Users/ozan.sazak/dev/xgotop
source env/bin/activate
pip install matplotlib numpy
```

## Usage

Basic usage with two metric files:

```bash
python plot_metrics.py --files "Label1:file1.json" "Label2:file2.json"
```

Specify custom output filename:

```bash
python plot_metrics.py --files "Regular:metrics_regular.json" "Web:metrics_web.json" --output comparison.png
```

Use different color palettes:

```bash
# Default vibrant palette
python plot_metrics.py --files "Regular:metrics.json" "Web:metrics_web.json"

# Dark cyberpunk theme with neon colors
python plot_metrics.py --files "Regular:metrics.json" "Web:metrics_web.json" --palette cyberpunk

# Clean brutalist theme
python plot_metrics.py --files "Regular:metrics.json" "Web:metrics_web.json" --palette brutalist
```

### Plot Modes

The script supports two visualization modes:

**All Metrics (default)**
- Creates a 2x2 grid showing all metrics (RPS, PPS, EWP, LAT)
- Each metric is plotted separately across all files

```bash
python plot_metrics.py --files "Regular:metrics.json" "Web:metrics_web.json" --mode all
```

**RPS vs PPS Comparison**
- Creates one subplot per JSON file
- Shows RPS and PPS on the same plot with area fill between them
- Makes it easy to visualize the gap between reads and processing

```bash
python plot_metrics.py --files "Regular:metrics.json" "Web:metrics_web.json" --mode rps-pps
```

### Examples

Compare regular vs web performance (all metrics):

```bash
python plot_metrics.py --files "Regular:metrics_2025-11-08-15-06-22.json" "Web:metrics_2025-11-08-15-09-19_web.json" --output metrics_comparison.png
```

RPS vs PPS comparison with area plot:

```bash
python plot_metrics.py --files "Regular:metrics_2025-11-08-15-06-22.json" "Web:metrics_2025-11-08-15-09-19_web.json" --mode rps-pps --output rps_pps_gap.png
```

RPS vs PPS with cyberpunk theme:

```bash
python plot_metrics.py --files "Regular:metrics_2025-11-08-15-06-22.json" "Web:metrics_2025-11-08-15-09-19_web.json" --mode rps-pps --palette cyberpunk
```

Compare multiple runs:

```bash
python plot_metrics.py --files "Run1:metrics1.json" "Run2:metrics2.json" "Run3:metrics3.json" --output multi_run_comparison.png
```

## JSON Format

The script expects JSON files with the following structure:

```json
{
  "rps": [29652, 29268, ...],     // Reads per second
  "pps": [29652, 29265, ...],     // Processed per second
  "ewp": [0, 0, 0, 1, ...],       // Events waiting to be processed
  "lat": [1080, 1085, ...],       // Latency in milliseconds
  "ts": [1762614321489590500, ...]  // Timestamps (not plotted)
}
```

## Design Elements

The neobrutalistic style includes:

- **Bold Colors**: Hot pink, orange, yellow, purple, and blue
- **Thick Borders**: 4-8px black borders throughout
- **Shadow Effects**: Offset shadows for depth
- **Geometric Patterns**: Subtle striped backgrounds
- **Corner Brackets**: Decorative elements on each subplot
- **Bold Typography**: Monospace fonts with heavy weights

### Color Palettes

**Vibrant (default)**
- Bright, high-contrast colors on light background
- Black borders and text
- Classic neobrutalist aesthetic

**Cyberpunk**
- Neon colors on near-black background
- Pink borders with white text
- Futuristic, high-tech appearance

**Brutalist**
- Muted primary colors on light gray
- Dark gray borders and text
- Clean, architectural feel

## Output

The script generates a single PNG image containing:
- 2x2 grid of metric plots
- Bold title at the top
- Individual legends for each subplot
- Thick black border around the entire figure

## Troubleshooting

If you encounter font cache warnings:
- This is normal on first run, matplotlib is building its font cache
- Subsequent runs will be faster

If plots appear empty:
- Check that your JSON files contain the expected metric keys
- Verify the data arrays are not empty

## Design Notes

The script includes several visual enhancements:
- **Dynamic shadows**: Shadow offsets are calculated based on data ranges for optimal visibility
- **Palette-aware styling**: Text and shadow colors automatically adjust for dark/light backgrounds
- **Cyberpunk theme**: Uses white text and shadows on dark backgrounds for visibility
- **Responsive layout**: Elements scale appropriately with data ranges

