set GOOS=linux
set GOARCH=amd64
go build -o ../build/prcs/coinpay_prcs_linux -ldflags="-w -s" ./main.go
go build -o "C:\Users\Karim\Documents\Sources\CoinPay-PaymentSystem\_ DOCKER\prcs\coinpay_prcs_linux" -ldflags="-w -s" ./main.go