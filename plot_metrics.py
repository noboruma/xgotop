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

LINESTYLES = [
    dot + line 
    for dot in ('.', 'o', '^', 'v', '+', 'x')
    for line in ('-', '--', '-.', ':')
]

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

def create_rps_pps_comparison(metrics_data, labels, output_path, palette_name='vibrant'):
    """Create RPS vs PPS comparison plots with area between them"""
    global COLORS
    COLORS = PALETTES.get(palette_name, PALETTES['vibrant'])
    setup_neobrutalistic_style()
    
    # Create figure with subplots (one per dataset)
    n_datasets = len(metrics_data)
    # Special handling for single dataset - use 2x1 layout
    if n_datasets == 1:
        fig = plt.figure(figsize=(12, 12))
        n_rows = 2
    else:
        fig = plt.figure(figsize=(12, 6 * n_datasets))
        n_rows = n_datasets
    
    # Add a bold title with shadow effect
    fig.suptitle('RPS vs PPS COMPARISON', 
                 fontsize=36, weight='black', y=0.98, color=COLORS['text'])
    
    # Create subplot for each dataset
    for idx, (data, label) in enumerate(zip(metrics_data, labels)):
        ax = plt.subplot(n_rows, 1, idx + 1)
        
        if 'rps' in data and 'pps' in data:
            # Create x-axis values
            x_values = np.arange(len(data['rps']))
            
            # Get RPS and PPS data
            rps_data = np.array(data['rps'])
            pps_data = np.array(data['pps'])
            
            # Add shadows
            shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
            shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
            x_offset = len(x_values) * 0.003
            y_offset_rps = (max(rps_data) - min(rps_data)) * 0.01
            y_offset_pps = (max(pps_data) - min(pps_data)) * 0.01
            
            # Plot shadows
            ax.plot(x_values + x_offset, rps_data - y_offset_rps,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            ax.plot(x_values + x_offset, pps_data - y_offset_pps,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            
            # Plot RPS and PPS
            rps_line = ax.plot(x_values, rps_data, 
                              color=COLORS['primary'], linewidth=4, 
                              label='RPS (Reads)', zorder=3)
            pps_line = ax.plot(x_values, pps_data, 
                              color=COLORS['secondary'], linewidth=4, 
                              label='PPS (Processed)', zorder=3)
            
            # Fill area between RPS and PPS
            ax.fill_between(x_values, rps_data, pps_data, 
                           alpha=0.3, color=COLORS['tertiary'], 
                           label='Read-Process Gap', zorder=2)
            
            # Styling
            ax.set_title(f'{label} - RPS vs PPS', 
                        fontsize=20, weight='black', pad=20, color=COLORS['text'])
            ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
            ax.set_ylabel('OPERATIONS/SEC', fontsize=14, weight='bold', color=COLORS['text'])
            ax.tick_params(colors=COLORS['text'], which='both')
            
            # Add grid for better readability
            ax.grid(True, alpha=0.3, color=COLORS['text'], linewidth=1, linestyle='--')
            
            # Legend
            legend = ax.legend(loc='upper right', frameon=True, 
                             fancybox=False, shadow=False,
                             edgecolor=COLORS['border'], 
                             facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                             prop={'weight': 'bold', 'size': 12})
            legend.get_frame().set_linewidth(3)
            
            # Set legend text color
            for text in legend.get_texts():
                text.set_color('black' if palette_name != 'cyberpunk' else 'white')
            
            # Add decorative corner brackets
            xlim = ax.get_xlim()
            ylim = ax.get_ylim()
            bracket_size = 0.03
            
            # Top-left corner bracket
            ax.plot([xlim[0], xlim[0] + (xlim[1]-xlim[0])*bracket_size], 
                   [ylim[1], ylim[1]], color=COLORS['border'], linewidth=6)
            ax.plot([xlim[0], xlim[0]], 
                   [ylim[1], ylim[1] - (ylim[1]-ylim[0])*bracket_size], 
                   color=COLORS['border'], linewidth=6)
            
            # Bottom-right corner bracket
            ax.plot([xlim[1] - (xlim[1]-xlim[0])*bracket_size, xlim[1]], 
                   [ylim[0], ylim[0]], color=COLORS['border'], linewidth=6)
            ax.plot([xlim[1], xlim[1]], 
                   [ylim[0], ylim[0] + (ylim[1]-ylim[0])*bracket_size], 
                   color=COLORS['border'], linewidth=6)
            
            # Make spines thicker
            for spine in ax.spines.values():
                spine.set_linewidth(4)
    
    # If single dataset, add summary statistics in second subplot
    if n_datasets == 1 and len(metrics_data) > 0:
        ax2 = plt.subplot(2, 1, 2)
        data = metrics_data[0]
        
        if 'rps' in data and 'pps' in data:
            rps_data = np.array(data['rps'])
            pps_data = np.array(data['pps'])
            
            # Calculate statistics
            avg_rps = np.mean(rps_data)
            avg_pps = np.mean(pps_data)
            max_rps = np.max(rps_data)
            max_pps = np.max(pps_data)
            min_rps = np.min(rps_data)
            min_pps = np.min(pps_data)
            avg_gap = np.mean(rps_data - pps_data)
            
            # Clear axis
            ax2.clear()
            ax2.set_xlim(0, 10)
            ax2.set_ylim(0, 10)
            ax2.axis('off')
            
            # Create neobrutalistic info boxes
            box_width = 4
            box_height = 1.5
            
            # Title box
            title_rect = patches.Rectangle((1, 8), 8, 1.5, 
                                         facecolor=COLORS['primary'], 
                                         edgecolor=COLORS['border'],
                                         linewidth=4)
            ax2.add_patch(title_rect)
            ax2.text(5, 8.75, 'PERFORMANCE SUMMARY', 
                    ha='center', va='center', fontsize=20, 
                    weight='black', color='white')
            
            # RPS Stats box
            rps_rect = patches.Rectangle((0.5, 5.5), box_width, box_height,
                                       facecolor=COLORS['secondary'],
                                       edgecolor=COLORS['border'],
                                       linewidth=3)
            ax2.add_patch(rps_rect)
            ax2.text(2.5, 6.8, 'RPS STATS', ha='center', va='center',
                    fontsize=14, weight='black', color='white')
            ax2.text(2.5, 6.2, f'AVG: {avg_rps:.2f}', ha='center', va='center',
                    fontsize=12, weight='bold', color='white')
            ax2.text(2.5, 5.8, f'MIN: {min_rps:.2f} | MAX: {max_rps:.2f}', 
                    ha='center', va='center', fontsize=10, weight='bold', color='white')
            
            # PPS Stats box
            pps_rect = patches.Rectangle((5.5, 5.5), box_width, box_height,
                                       facecolor=COLORS['tertiary'],
                                       edgecolor=COLORS['border'],
                                       linewidth=3)
            ax2.add_patch(pps_rect)
            ax2.text(7.5, 6.8, 'PPS STATS', ha='center', va='center',
                    fontsize=14, weight='black', color=COLORS['border'])
            ax2.text(7.5, 6.2, f'AVG: {avg_pps:.2f}', ha='center', va='center',
                    fontsize=12, weight='bold', color=COLORS['border'])
            ax2.text(7.5, 5.8, f'MIN: {min_pps:.2f} | MAX: {max_pps:.2f}', 
                    ha='center', va='center', fontsize=10, weight='bold', color=COLORS['border'])
            
            # Gap Analysis box
            gap_rect = patches.Rectangle((2, 3), 6, box_height,
                                       facecolor=COLORS['quaternary'],
                                       edgecolor=COLORS['border'],
                                       linewidth=3)
            ax2.add_patch(gap_rect)
            ax2.text(5, 4.3, 'READ-PROCESS GAP', ha='center', va='center',
                    fontsize=14, weight='black', color=COLORS['border'])
            ax2.text(5, 3.5, f'AVERAGE GAP: {avg_gap:.2f} ops/sec', 
                    ha='center', va='center', fontsize=12, weight='bold', color=COLORS['border'])
            
            # Add decorative elements
            # Shadow effects
            shadow_offset = 0.1
            for rect in [title_rect, rps_rect, pps_rect, gap_rect]:
                shadow = patches.Rectangle((rect.get_x() + shadow_offset, 
                                          rect.get_y() - shadow_offset),
                                         rect.get_width(), rect.get_height(),
                                         facecolor='black', alpha=0.3, zorder=0)
                ax2.add_patch(shadow)
    
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


def create_metric_plots(metrics_data, labels, output_path, palette_name='vibrant'):
    """Create neobrutalistic plots with RPS vs PPS for each dataset and other metrics"""
    global COLORS
    COLORS = PALETTES.get(palette_name, PALETTES['vibrant'])
    setup_neobrutalistic_style()
    
    # Create figure with 2x2 grid
    fig = plt.figure(figsize=(20, 14))
    
    # Add a bold title with shadow effect
    fig.suptitle('PERFORMANCE METRICS ANALYSIS', 
                 fontsize=36, weight='black', y=0.98, color=COLORS['text'])
    
    # Special handling for single dataset
    if len(metrics_data) == 1:
        # Single dataset: RPS vs PPS spans two columns
        data = metrics_data[0]
        label = labels[0]
        ax = plt.subplot(2, 2, (1, 2))  # Span columns 1 and 2
        
        if 'rps' in data and 'pps' in data:
            # Create x-axis values
            x_values = np.arange(len(data['rps']))
            
            # Get RPS and PPS data
            rps_data = np.array(data['rps'])
            pps_data = np.array(data['pps'])
            
            # Add shadows
            shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
            shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
            x_offset = len(x_values) * 0.003
            y_offset_rps = (max(rps_data) - min(rps_data)) * 0.01
            y_offset_pps = (max(pps_data) - min(pps_data)) * 0.01
            
            # Plot shadows
            ax.plot(x_values + x_offset, rps_data - y_offset_rps,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            ax.plot(x_values + x_offset, pps_data - y_offset_pps,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            
            # Plot RPS and PPS
            ax.plot(x_values, rps_data, 
                   color=COLORS['primary'], linewidth=4, 
                   label='RPS (Reads)', zorder=3)
            ax.plot(x_values, pps_data, 
                   color=COLORS['secondary'], linewidth=4, 
                   label='PPS (Processed)', zorder=3)
            
            # Fill area between RPS and PPS
            ax.fill_between(x_values, rps_data, pps_data, 
                           alpha=0.3, color=COLORS['tertiary'], 
                           label='Read-Process Gap', zorder=2)
            
            # Styling
            ax.set_title(f'{label} - RPS vs PPS', 
                        fontsize=18, weight='black', pad=20, color=COLORS['text'])
            ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
            ax.set_ylabel('OPERATIONS/SEC', fontsize=14, weight='bold', color=COLORS['text'])
            ax.tick_params(colors=COLORS['text'], which='both')
            
            # Double the x-axis ticks since plot spans two columns
            ax.xaxis.set_major_locator(plt.MaxNLocator(integer=True, nbins=20))
            
            # Legend
            legend = ax.legend(loc='upper right', frameon=True, 
                             fancybox=False, shadow=False,
                             edgecolor=COLORS['border'], 
                             facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                             prop={'weight': 'bold', 'size': 10})
            legend.get_frame().set_linewidth(3)
            
            # Set legend text color
            for text in legend.get_texts():
                text.set_color('black' if palette_name != 'cyberpunk' else 'white')
    else:
        # Multiple datasets: First two plots are RPS vs PPS for each dataset
        for idx, (data, label) in enumerate(zip(metrics_data[:2], labels[:2])):
            ax = plt.subplot(2, 2, idx + 1)
            
            if 'rps' in data and 'pps' in data:
                # Create x-axis values
                x_values = np.arange(len(data['rps']))
                
                # Get RPS and PPS data
                rps_data = np.array(data['rps'])
                pps_data = np.array(data['pps'])
                
                # Add shadows
                shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
                shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
                x_offset = len(x_values) * 0.003
                y_offset_rps = (max(rps_data) - min(rps_data)) * 0.01
                y_offset_pps = (max(pps_data) - min(pps_data)) * 0.01
                
                # Plot shadows
                ax.plot(x_values + x_offset, rps_data - y_offset_rps,
                       color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
                ax.plot(x_values + x_offset, pps_data - y_offset_pps,
                       color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
                
                # Plot RPS and PPS
                ax.plot(x_values, rps_data, 
                       color=COLORS['primary'], linewidth=4, 
                       label='RPS (Reads)', zorder=3)
                ax.plot(x_values, pps_data, 
                       color=COLORS['secondary'], linewidth=4, 
                       label='PPS (Processed)', zorder=3)
                
                # Fill area between RPS and PPS
                ax.fill_between(x_values, rps_data, pps_data, 
                               alpha=0.3, color=COLORS['tertiary'], 
                               label='Read-Process Gap', zorder=2)
                
                # Styling
                ax.set_title(f'{label} - RPS vs PPS', 
                            fontsize=18, weight='black', pad=20, color=COLORS['text'])
                ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
                ax.set_ylabel('OPERATIONS/SEC', fontsize=14, weight='bold', color=COLORS['text'])
                ax.tick_params(colors=COLORS['text'], which='both')
                
                # Legend
                legend = ax.legend(loc='upper right', frameon=True, 
                                 fancybox=False, shadow=False,
                                 edgecolor=COLORS['border'], 
                                 facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                                 prop={'weight': 'bold', 'size': 10})
                legend.get_frame().set_linewidth(3)
                
                # Set legend text color
                for text in legend.get_texts():
                    text.set_color('black' if palette_name != 'cyberpunk' else 'white')
    
    # Plot 3: Events Waiting to be Processed (EWP)
    ax = plt.subplot(2, 2, 3)
    for i, (data, label) in enumerate(zip(metrics_data, labels)):
        if 'ewp' in data:
            color = list(COLORS.values())[i % len(COLORS)]
            linestyle = LINESTYLES[i % len(LINESTYLES)]
            x_values = np.arange(len(data['ewp']))
            
            # Shadow
            shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
            shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
            x_offset = len(x_values) * 0.003
            y_offset = max(1, (max(data['ewp']) - min(data['ewp'])) * 0.01)
            
            ax.plot(x_values + x_offset, np.array(data['ewp']) - y_offset, linestyle,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            
            # Main plot
            ax.plot(x_values, data['ewp'], linestyle,
                   color=color, linewidth=4, label=label, zorder=2)
    
    ax.set_title('EVENTS WAITING TO BE PROCESSED', 
                fontsize=18, weight='black', pad=20, color=COLORS['text'])
    ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
    ax.set_ylabel('COUNT', fontsize=14, weight='bold', color=COLORS['text'])
    ax.tick_params(colors=COLORS['text'], which='both')
    
    # Legend
    legend = ax.legend(loc='upper right', frameon=True, 
                     fancybox=False, shadow=False,
                     edgecolor=COLORS['border'], 
                     facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                     prop={'weight': 'bold', 'size': 10})
    legend.get_frame().set_linewidth(3)
    for text in legend.get_texts():
        text.set_color('black' if palette_name != 'cyberpunk' else 'white')
    
    # Plot 4: Latency
    ax = plt.subplot(2, 2, 4)
    for i, (data, label) in enumerate(zip(metrics_data, labels)):
        if 'lat' in data:
            color = list(COLORS.values())[i % len(COLORS)]
            linestyle = LINESTYLES[i % len(LINESTYLES)]
            x_values = np.arange(len(data['lat']))
            
            # Shadow
            shadow_color = 'white' if palette_name == 'cyberpunk' else 'black'
            shadow_alpha = 0.2 if palette_name == 'cyberpunk' else 0.3
            x_offset = len(x_values) * 0.003
            y_offset = (max(data['lat']) - min(data['lat'])) * 0.01
            
            ax.plot(x_values + x_offset, np.array(data['lat']) - y_offset, linestyle,
                   color=shadow_color, linewidth=5, alpha=shadow_alpha, zorder=1)
            
            # Main plot
            ax.plot(x_values, data['lat'], linestyle,
                   color=color, linewidth=4, label=label, zorder=2)
    
    ax.set_title('LATENCY (ms)', 
                fontsize=18, weight='black', pad=20, color=COLORS['text'])
    ax.set_xlabel('SAMPLE', fontsize=14, weight='bold', color=COLORS['text'])
    ax.set_ylabel('MILLISECONDS', fontsize=14, weight='bold', color=COLORS['text'])
    ax.tick_params(colors=COLORS['text'], which='both')
    
    # Legend
    legend = ax.legend(loc='upper right', frameon=True, 
                     fancybox=False, shadow=False,
                     edgecolor=COLORS['border'], 
                     facecolor='white' if palette_name != 'cyberpunk' else COLORS['border'],
                     prop={'weight': 'bold', 'size': 10})
    legend.get_frame().set_linewidth(3)
    for text in legend.get_texts():
        text.set_color('black' if palette_name != 'cyberpunk' else 'white')
    
    # Add decorative elements to all subplots
    if len(metrics_data) == 1:
        # For single dataset: RPS/PPS spans (1,2), EWP is 3, Latency is 4
        subplot_positions = [(1, 2), 3, 4]
        for pos in subplot_positions:
            if isinstance(pos, tuple):
                ax = plt.subplot(2, 2, pos)
            else:
                ax = plt.subplot(2, 2, pos)
            
            # Corner brackets
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
                
            # Add background pattern (diagonal stripes)
            for i in range(0, 100, 10):
                ax.axhspan(ax.get_ylim()[0] + i * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                          ax.get_ylim()[0] + (i + 5) * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                          facecolor='white', alpha=0.1, zorder=0)
    else:
        # For multiple datasets: regular 2x2 grid
        for idx in range(1, 5):
            ax = plt.subplot(2, 2, idx)
        
            # Corner brackets
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
                
            # Add background pattern (diagonal stripes)
            for i in range(0, 100, 10):
                ax.axhspan(ax.get_ylim()[0] + i * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                          ax.get_ylim()[0] + (i + 5) * (ax.get_ylim()[1] - ax.get_ylim()[0]) / 100,
                          facecolor='white', alpha=0.1, zorder=0)
    
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
    parser.add_argument('--mode', '-m', default='all',
                       choices=['all', 'rps-pps'],
                       help='Plot mode: all metrics or RPS vs PPS comparison (default: all)')
    
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
    
    # Create plots based on mode
    if args.mode == 'rps-pps':
        create_rps_pps_comparison(metrics_data, labels, args.output, args.palette)
    else:
        create_metric_plots(metrics_data, labels, args.output, args.palette)

if __name__ == "__main__":
    main()
