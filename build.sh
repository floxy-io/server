cp -R internal/home/assets build/assets
find build -type f | xargs sed -i  's/http:\/\/localhost:8080/https:\/\/floxy\.io/g'
go build -ldflags="-X 'github.com/danielsussa/floxy/internal/home.AssetsPath=build/assets'" -o build/floxy internal/cmd/main.go