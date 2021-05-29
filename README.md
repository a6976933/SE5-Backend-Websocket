# SE5-Backend-Websocket
se5-backend-websocket server


## 目前進度
新增JWT 加密方法為RS256 
### 生成Key的方式
生成的Key為PEM檔 因為生成其它格式解碼較為麻煩 採用pbcs8格式
#### 生成private key
```
openssl genrsa -out key.pem 2048
```
#### 生成public key
```
openssl rsa -in key.pem -pubout -out key.pem.pub
```

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

