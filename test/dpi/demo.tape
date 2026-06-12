# VHS tape — alternative GIF recorder for the DPI-evasion demo.
#   Install: https://github.com/charmbracelet/vhs
#   Prereqs: docker + harness images built (test/dpi/build.sh)
#   Generate: vhs test/dpi/demo.tape   ->   test/dpi/demo.gif
#
# Note: VHS records wall-clock time, so the docker bring-up/tear-down between
# scenarios shows as live spinner time. For a tighter result, prefer the
# asciinema + agg flow documented in demo.sh (it auto-collapses idle waits).

Output test/dpi/demo.gif

Set Shell bash
Set FontSize 20
Set Width 1120
Set Height 640
Set Padding 24
Set Theme "Dracula"

Hide
Type "bash test/dpi/demo.sh" Enter
Show
# Cover both docker scenarios; trim afterward with gifsicle if needed.
Sleep 120s
