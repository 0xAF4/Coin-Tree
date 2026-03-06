set GOOS=windows
set GOARCH=amd64
go build -o ../build/auth/coinpay_prcs.exe -ldflags="-w -s" ./main.go