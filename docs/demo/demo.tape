# prism demo recording script for vhs (https://github.com/charmbracelet/vhs)
#
# Usage:
#   export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
#   vhs docs/demo/demo.tape
#
# Output:
#   docs/demo/demo.gif

Output docs/demo/demo.gif

Set Shell fish
#Set FontFamily "Noto Sans Mono CJK JP"
Set FontFamily "Monoflow"
Set Theme { "name": "Whimsy", "black": "#535178", "red": "#ef6487", "green": "#5eca89", "yellow": "#fdd877", "blue": "#65aef7", "magenta": "#aa7ff0", "cyan": "#43c1be", "white": "#ffffff", "brightBlack": "#535178", "brightRed": "#ef6487", "brightGreen": "#5eca89", "brightYellow": "#fdd877", "brightBlue": "#65aef7", "brightMagenta": "#aa7ff0", "brightCyan": "#43c1be", "brightWhite": "#ffffff", "background": "#29283b", "foreground": "#b3b0d6", "selection": "#3d3c58", "cursor": "#b3b0d6" }
Set FontSize 16
Set Width 1280
Set Height 720
#Set Theme "Dracula"
Set PlaybackSpeed 1.0
Set TypingSpeed 40ms
Set Padding 0
Set WindowBar Colorful

# Analyze a real PR
Type "prism analyze https://github.com/hidetzu/prism/pull/1 --format markdown"
Sleep 500ms
Enter
Sleep 6s

# Clear and show prompt generation
Ctrl+L
Sleep 500ms
Type "prism prompt https://github.com/hidetzu/prism/pull/1 --mode detailed | head -40"
Sleep 500ms
Enter
Sleep 6s

Sleep 2s
