compile_time=$(shell date +%Y/%m/%d-%H:%M:%S)
version=$(shell go version | awk '{print $$3}')
branch=$(shell git branch | grep '*' | awk '{print $$2}')
commit=$(shell git log | grep commit | head -n 1 | awk '{print $$2}')




debug:
	@go build -race -gcflags="-N -l" -ldflags="-X github.com/Hurricanezwf/rate-limiter/version.GoVersion=$(version) -X github.com/Hurricanezwf/rate-limiter/version.Branch=$(branch) -X github.com/Hurricanezwf/rate-limiter/version.Commit=$(commit) -X github.com/Hurricanezwf/rate-limiter/version.CompileTime=$(compile_time)"



release:
	@export CGO_ENABLED=0 && go build -ldflags="-w -s -X github.com/Hurricanezwf/rate-limiter/version.GoVersion=$(version) -X github.com/Hurricanezwf/rate-limiter/version.Branch=$(branch) -X github.com/Hurricanezwf/rate-limiter/version.Commit=$(commit) -X github.com/Hurricanezwf/rate-limiter/version.CompileTime=$(compile_time)"



clean:
	@rm  -f ./rate-limiter
