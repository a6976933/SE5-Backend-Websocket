# SE5-Backend-Websocket
se5-backend-websocket server


## 目前進度
JWT 加密方法為HS256
完成ORM存取
完成通知系統(或剩下要討論的地方?)
尚未tested
剩餘事項: 去Commit front-end

## 測試的安裝步驟
### go install
https://golang.org/doc/install

### git add remote
```
git remote add origin git@github.com:a6976933/SE5-Backend-Websocket.git
```
### git pull
```
git pull origin master
```
### go path setting
```
export PATH=$PATH:yourGoInstallPath/go/bin
```
### download package
```
cd projectdir
```
```
go mod download
```
### build
```
go build main.go
```
### run
```
./main
```

