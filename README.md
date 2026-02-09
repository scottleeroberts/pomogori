# pomogori

A simple, powerful pomodoro timer for the terminal with desktop notifications and i3 window manager integration.

## Features

- Work and break session tracking with customizable durations
- Desktop notifications on session completion (Linux, macOS, Windows)
- Live countdown display with progress bar
- Pause and resume functionality
- Session history and statistics
- i3 window manager integration with status bar display
- Configurable default durations
- Persistent state across terminal sessions

## Installation

### Build from source

```bash
git clone https://github.com/yourusername/pomogori.git
cd pomogori
go build -o pomogori main.go
```

### Install to system

```bash
# Copy to a directory in your PATH
sudo cp pomogori /usr/local/bin/
# Or for user-only install
cp pomogori ~/.local/bin/
```

## Basic Usage

```bash
# Start a work session (default: 25 minutes)
pomogori work

# Start a break session (default: 5 minutes)
pomogori break

# Custom duration (in minutes)
pomogori work -d 50
pomogori break -d 15

# Check remaining time
pomogori status

# Watch with live countdown
pomogori watch

# Pause/resume
pomogori pause
pomogori resume

# Stop current timer
pomogori stop

# View session statistics
pomogori stats
```

## Configuration

Edit `~/.pomogori/config` to customize default durations:

```
work_minutes = 25
break_minutes = 5
```

## i3 Window Manager Integration

pomogori integrates seamlessly with i3 to provide visual timer feedback in your status bar and quick keyboard access to all timer functions.

### i3blocks Status Bar Integration

The `pomogori i3blocks` command outputs formatted status for i3blocks:
- 🍅 MM:SS - Active work session
- ☕ MM:SS - Active break session
- ⏸️ - Paused indicator
- (no output when stopped/complete)

#### Setup i3blocks

1. Add pomogori to your i3blocks configuration (`~/.config/i3/i3blocks.conf`):

```ini
[pomo]
command=/path/to/pomogori i3blocks
interval=1
```

2. Restart i3blocks:
```bash
killall i3blocks
i3-msg restart
```

The timer will now appear in your status bar and update every second.

### i3 Keybinding Integration

Add a pomodoro mode to your i3 config (`~/.config/i3/config`) for quick keyboard access:

```
# Pomodoro management mode
set $pomo Pomodoro: (w)ork (b)reak (p)ause (r)esume (s)top (t)atus
mode "$pomo" {
    bindsym w exec --no-startup-id pomogori work, mode "default"
    bindsym b exec --no-startup-id pomogori break, mode "default"
    bindsym p exec --no-startup-id pomogori pause, mode "default"
    bindsym r exec --no-startup-id pomogori resume, mode "default"
    bindsym s exec --no-startup-id pomogori stop, mode "default"
    bindsym t exec --no-startup-id alacritty -e pomogori watch, mode "default"

    bindsym Return mode "default"
    bindsym Escape mode "default"
}

# Trigger with $mod+g
bindsym $mod+g mode "$pomo"
```

**Usage:**
- Press `$mod+g` to enter pomodoro mode
- Press `w` to start a work session
- Press `b` to start a break
- Press `p` to pause, `r` to resume
- Press `s` to stop the timer
- Press `t` to open watch mode in a terminal
- Press `Escape` or `Return` to exit mode

Reload i3 config: `$mod+Shift+r`

### Complete i3 Workflow Example

1. Press `$mod+g` then `w` - Start 25-minute work session
2. Status bar shows: 🍅 24:58
3. Work on your task - timer counts down in status bar
4. Need a quick break? `$mod+g` then `p` - Pause timer
5. Status bar shows: 🍅 15:23 ⏸️
6. Back to work? `$mod+g` then `r` - Resume timer
7. Timer completes - Desktop notification appears
8. Press `$mod+g` then `b` - Start 5-minute break
9. Status bar shows: ☕ 04:58

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `work` | `w` | Start a work session |
| `break` | `b` | Start a break session |
| `status` | `s` | Show current timer status |
| `pause` | `p` | Pause the current timer |
| `resume` | `r` | Resume a paused timer |
| `watch` | - | Watch timer with live countdown |
| `stop` | - | Stop and discard current timer |
| `stats` | - | Show session statistics |
| `config` | - | Show configuration |
| `i3blocks` | `i3b` | Output status for i3blocks |
| `help` | `h` | Show help message |
| `version` | `v` | Show version |

## Data Storage

pomogori stores its data in `~/.pomogori/`:
- `config` - Configuration file
- `state` - Current timer state
- `history` - Completed session history

## Requirements

- Go 1.16+ (for building)
- Linux: `notify-send` (usually pre-installed)
- macOS: `osascript` (built-in)
- Windows: PowerShell (built-in)

## License

MIT

## Author

Built for focused work sessions and integrated deeply with the i3 window manager workflow.
