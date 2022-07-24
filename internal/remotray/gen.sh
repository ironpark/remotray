GOOS=darwin go build -o ../../tray .
GOOS=windows go build  -ldflags -H=windowsgui -o  ../../tray.exe .