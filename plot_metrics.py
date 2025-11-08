#!/usr/bin/env python3
"""
Neobrutalistic metric visualization script
Generates plots from multiple metric JSON files with bold, geometric styling
"""

import json
import argparse
import matplotlib.pyplot as plt
import matplotlib.patches as patches
import numpy as np
from pathlib import Path
import sys

# Neobrutalistic color palettes
PALETTES = {
    'vibrant': {
        'primary': '#FF006E',      # Hot pink
        'secondary': '#FB5607',    # Orange
        'tertiary': '#FFBE0B',     # Yellow
        'quaternary': '#8338EC',   # Purple
        'quinary': '#3A86FF',      # Blue
        'background': '#F3F3F3',   # Light gray
        'text': '#000000',         # Black
        'border': '#000000'        # Black borders
    },
    'cyberpunk': {
        'primary': '#FF0080',      # Neon pink
        'secondary': '#00FF88',    # Neon green
        'tertiary': '#00E5FF',     # Cyan
        'quaternary': '#FFD700',   # Gold
        'quinary': '#FF00FF',      # Magenta
        'background': '#0A0A0A',   # Near black
        'text': '#FFFFFF',         # White
        'border': '#FF0080'        # Neon pink borders
    },
    'brutalist': {
        'primary': '#D32F2F',      # Red
        'secondary': '#388E3C',    # Green
        'tertiary': '#1976D2',     # Blue
        'quaternary': '#FBC02D',   # Yellow
        'quinary': '#7B1FA2',      # Purple
        'background': '#ECEFF1',   # Light gray
        'text': '#212121',         # Dark gray
        'border': '#212121'        # Dark gray borders
    }
}

# Default colors
COLORS = PALETTES['vibrant']

def load_metrics(json_path):
    """Load metrics from JSON file"""
    with open(json_path, 'r') as f:
        data = json.load(f)
    return data

def setup_neobrutalistic_style():
    """Configure matplotlib for neobrutalistic aesthetic"""
    plt.rcParams.update({
        'font.family': 'monospace',
        'font.weight': 'bold',
        'font.size': 12,
        'axes.linewidth': 4,
        'axes.edgecolor': COLORS['border'],
        'axes.facecolor': COLORS['background'],
        'figure.facecolor': COLORS['background'],
        'lines.linewidth': 3,
        'xtick.major.width': 3,
        'ytick.major.width': 3,
        'xtick.major.size': 8,
        'ytick.major.size': 8,
        'axes.grid': False,
    })

def create_metric_plots(metrics_data, labels, output_path, palette_name='vibrant'):
    """Create neobrutalistic plots for all metrics"""
    global COLORS
    COLORS = PALETTES.get(palette_name, PALETTES['vibrant'])
    setup_neobrutalistic_style()
    
    # Determine which metrics to plot (skip 'ts' as it's just timestamps)
    metric_names = ['rps', 'pps', 'ewp', 'lat']
    metric_titles = {
        'rps': 'READS PER SECOND',
        'pps': 'PROCESSED PER SECOND', 
        'ewp': 'EVENTS WAITING TO BE PROCESSED',
        'lat': 'LATENCY (ms)'
    }
    
    # ASCII art decorations
    ascii_patterns = ['///', '\\\\\\', '|||', '---', '+++', 'xxx']
    
    # Create figure with subplots
    fig = plt.figure(figsize=(20, 14))
    
    # Add a bold title with shadow effect
    fig.suptitle('PERFORMANCE METRICS ANALYSIS', 
                 fontsize=36, weight='black', y=0.98, color=COLORS['text'])
    
    # Create 2x2 grid
    for idx, metric in enumerate(metric_names):
        ax = plt.subplot(2, 2, idx + 1)
        
        # Plot each dataset
        for i, (data, label) in enumerate(zip(metrics_data, labels)):
            if metric in data:
                # Use different colors for each dataset
                color = list(COLORS.values())[i % len(COLORS)]
                
                # Create x-axis values (sample index)
                x_values = np.arange(len(data[metric]))
                
                # Add shadow effect - plot same data offset
                # Calculate dynamic offset based on data range
                data_range = max(data[metric]) - min(data[metric])
                y_offset = data_range * 0.01  # 1% of data range
                x_offset = len(x_values) * 0.003  # 0.3% of x range
                
                # Use appropriate shadow color based on palette
                shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
                shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
                
                ax.plot(x_values + x_offset, np.array(data[metric]) - y_offset, 
                       color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
                
                # Main plot line
                ax.plot(x_values, data[metric], 
                       color=color, linewidth=4, label=label, zorder=2)
        
        # Styling
        ax.set_title(metric_titles.get(metric, metric.upper()), 
                    fontsize=18, weight='black', pad=20, color=COLORS['text'])
        ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
        ax.set_ylabel('VALUE', fontsize=14, weight='bold', color=COLORS['text'])
        
        # Set tick colors
        ax.tick_params(colors=COLORS['text'], which='both')
        
        # Add background pattern (diagonal stripes)
        for i in range(0, 100, 10):
            ax.axhspan(ax.get_ylim()[0] + i * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                      ax.get_ylim()[0] + (i + 5) * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                      facecolor='white', alpha=0.1, zorder=0)
        
        # Bold legend with border
        legend = ax.legend(loc='upper right', frameon=True, 
                          fancybox=False, shadow=False,
                          edgecolor=COLORS['border'], 
                          facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                          prop={'weight': 'bold', 'size': 10})
        legend.get_frame().set_linewidth(3)
        
        # Set legend text color
        for text in legend.get_texts():
            text.set_color('black' if palette_name != 'cyberpunk' else 'white')
        
        # Add decorative elements - corner brackets
        xlim = ax.get_xlim()
        ylim = ax.get_ylim()
        bracket_size = 0.05
        
        # Top-left corner bracket
        ax.plot([xlim[0], xlim[0] + (xlim[1]-xlim[0])*bracket_size], 
               [ylim[1], ylim[1]], color=COLORS['border'], linewidth=6)
        ax.plot([xlim[0], xlim[0]], 
               [ylim[1], ylim[1] - (ylim[1]-ylim[0])*bracket_size], color=COLORS['border'], linewidth=6)
        
        # Bottom-right corner bracket
        ax.plot([xlim[1] - (xlim[1]-xlim[0])*bracket_size, xlim[1]], 
               [ylim[0], ylim[0]], color=COLORS['border'], linewidth=6)
        ax.plot([xlim[1], xlim[1]], 
               [ylim[0], ylim[0] + (ylim[1]-ylim[0])*bracket_size], color=COLORS['border'], linewidth=6)
        
        # Make spines thicker
        for spine in ax.spines.values():
            spine.set_linewidth(4)
    
    # Adjust layout
    plt.tight_layout()
    
    # Add decorative border around entire figure
    border_ax = fig.add_subplot(111, frameon=False)
    border_ax.tick_params(labelcolor='none', top=False, bottom=False, 
                         left=False, right=False)
    for spine in border_ax.spines.values():
        spine.set_linewidth(8)
        spine.set_edgecolor(COLORS['border'])
    
    # Save figure
    plt.savefig(output_path, dpi=300, bbox_inches='tight', 
                facecolor=COLORS['background'], edgecolor=COLORS['border'])
    print(f"Plot saved to: {output_path}")

def main():
    parser = argparse.ArgumentParser(
        description='Generate neobrutalistic metric plots from JSON files')
    parser.add_argument('--files', nargs='+', required=True,
                       help='JSON files with metrics, format: label:path')
    parser.add_argument('--output', '-o', default='metrics_plot.png',
                       help='Output PNG file path (default: metrics_plot.png)')
    parser.add_argument('--palette', '-p', default='vibrant',
                       choices=['vibrant', 'cyberpunk', 'brutalist'],
                       help='Color palette to use (default: vibrant)')
    
    args = parser.parse_args()
    
    # Parse files and labels
    metrics_data = []
    labels = []
    
    for file_spec in args.files:
        if ':' not in file_spec:
            print(f"Error: File spec must be in format 'label:path', got: {file_spec}")
            sys.exit(1)
        
        label, path = file_spec.split(':', 1)
        
        if not Path(path).exists():
            print(f"Error: File not found: {path}")
            sys.exit(1)
        
        try:
            data = load_metrics(path)
            metrics_data.append(data)
            labels.append(label)
            print(f"Loaded metrics from {path} with label '{label}'")
        except Exception as e:
            print(f"Error loading {path}: {e}")
            sys.exit(1)
    
    # Create plots
    create_metric_plots(metrics_data, labels, args.output, args.palette)

if __name__ == "__main__":
    main()
