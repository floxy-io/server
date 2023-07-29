rm -rf build/assets
cp -rfi internal/home/assets build/assets
find build -type f | xargs sed -i  's/http:\/\/localhost:8080/https:\/\/floxy\.io/g'
go build -ldflags="
  -X 'github.com/danielsussa/floxy/internal/home.AssetsPath=build/assets'
  -X 'github.com/danielsussa/floxy/internal/infra/compiler.CustomGoPath=/home/danielsussa/go'
  -X 'github.com/danielsussa/floxy/internal/infra/compiler.CustomPath=/home/danielsussa/go/src/github.com/danielsussa/floxy/internal/cook/cook.go'" \
   -o build/floxy cmd/main.go