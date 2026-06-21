# Regenerates docs/demo.gif with charm vhs (https://github.com/charmbracelet/vhs).
#
#   go build -o /tmp/mctop .
#   go run ./examples/demoserver &        # serves :8080
#   alias mctop=/tmp/mctop
#   vhs docs/demo.tape
#
# On a headless or root environment set VHS_NO_SANDBOX=1 first.

Output "docs/demo.gif"
Set Shell "zsh"
Set FontSize 15
Set Width 1080
Set Height 640
Set Padding 24

Hide
Type "PS1='$ '"
Enter
Type "clear"
Enter
Show

Sleep 700ms
Type "mctop http://localhost:8080/mcp"
Sleep 400ms
Enter
Sleep 2s

Down
Sleep 700ms
Up
Sleep 500ms
Down
Sleep 600ms
Enter
Sleep 900ms
Type "hello, terminal"
Sleep 800ms
Enter
Sleep 1800ms

Type "T"
Sleep 1600ms
Down@350ms 6
Sleep 1400ms
Escape
Sleep 600ms
Type "q"
Sleep 800ms
