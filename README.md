# go-polykit

**go-polykit** adalah pustaka arsitektur Golang kustom yang dirancang secara khusus untuk membangun *microservices* modern yang tangguh. Pustaka ini sepenuhnya **protocol-agnostic** dan **framework-agnostic**.

## 🌟 Mengapa Menggunakan go-polykit?

Mengadopsi pola "Endpoint" yang dipopulerkan oleh Go-kit, library ini memisahkan secara tegas antara *domain business logic* dan *transport layer* (HTTP, gRPC, WebSocket, GraphQL).

### Manfaat Utama:
1. **Performa Tinggi (Zero Reflection)**: Tidak seperti Go-kit yang sangat bergantung pada `interface{}` abstrak yang rumit, `go-polykit` mendelegasikan parsing request dan response kepada framework spesifik (Fiber/Mux/gRPC) dan sangat memanfaatkan *type assertion* langsung. Hal ini menjamin latensi yang sangat rendah dan meminimalisir beban *Garbage Collector*.
2. **Fleksibilitas Protokol**: Satu *Endpoint* (fungsi bisnis murni) dapat di-serve secara serentak ke berbagai transport layer:
   - GoFiber HTTP (berbasis `fasthttp`)
   - Gorilla Mux HTTP (berbasis `net/http`)
   - WebSocket (GoFiber & Gorilla)
   - gRPC
   - GraphQL
3. **Context Propagation & Standarisasi**: Secara otomatis mengambil token JWT atau kredensial Basic Auth dari HTTP Header dan menyuntikkannya ke dalam Context. Pustaka ini juga memperkenalkan mekanisme balasan (`StandardResponse`) yang konsisten agar mudah dikonsumsi *client*.
4. **Resilient Client**: Menyediakan fungsi klien (*HTTP, Fiber, gRPC*) yang dibungkus selayaknya *Endpoint* lokal. Klien secara otomatis mengubah error infrastruktur/jaringan (*timeout / connection refused*) menjadi `StandardResponse` (Kode 99), mencegah aplikasi mengalami *crash* mendadak (*Graceful Degradation*).

## 📦 Instalasi

Pastikan Anda menggunakan Go versi 1.21 ke atas.
```bash
go get github.com/addpur/go-polykit
```

## 🚀 Cara Penggunaan

### 1. Membuat Business Logic (Endpoint)
Pisahkan logika bisnis Anda ke dalam sebuah Endpoint.

```go
import "github.com/addpur/go-polykit"

type HelloReq struct { Name string }

func makeHelloEndpoint() polykit.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(HelloReq)
        userID := ctx.Value("user_id")
        return polykit.StandardResponse{
            ResponseCode: "00",
            Message: "Success",
            Data: "Hello " + req.Name + " (UserID: " + fmt.Sprintf("%v", userID) + ")",
        }, nil
    }
}
```

### 2. Inisialisasi Zap Logger

```go
import (
    "go.uber.org/zap"
    "github.com/addpur/go-polykit/pkg/polykit/logger"
)

zapLogger, _ := zap.NewDevelopment()
defer zapLogger.Sync()
zapLog := logger.NewLogger(zapLogger.Sugar())
```

### 3. Merangkai Middleware (Chain)

`go-polykit` menyediakan tiga middleware utama yang dapat dirangkai menggunakan `polykit.Chain`:

| Middleware | Paket | Fungsi |
|---|---|---|
| `TracingMiddleware` | `pkg/polykit/telemetry` | Memulai & mengakhiri OpenTelemetry span |
| `LoggingMiddleware` | `pkg/polykit/logger` | Mencatat request, response, durasi & error via Zap |
| `JWTAuthMiddleware` | `go-polykit` (core) | Validasi JWT Bearer token dari header |
| `BasicAuthMiddleware` | `go-polykit` (core) | Validasi HTTP Basic Auth dari header |

**Urutan eksekusi yang direkomendasikan:**

```
TracingMiddleware → LoggingMiddleware → JWTAuthMiddleware/BasicAuthMiddleware → Endpoint
```

**Contoh chain dengan JWT Auth:**
```go
import (
    "github.com/addpur/go-polykit"
    "github.com/addpur/go-polykit/pkg/polykit/logger"
    "github.com/addpur/go-polykit/pkg/polykit/telemetry"
)

endpoint := polykit.Chain(
    telemetry.TracingMiddleware("my-endpoint"),
    logger.LoggingMiddleware(zapLog, "my-endpoint"),
    polykit.JWTAuthMiddleware("my-secret-key"),
)(makeHelloEndpoint())
```

**Contoh chain dengan Basic Auth:**
```go
endpoint := polykit.Chain(
    telemetry.TracingMiddleware("my-endpoint"),
    logger.LoggingMiddleware(zapLog, "my-endpoint"),
    polykit.BasicAuthMiddleware("admin", "s3cr3t"),
)(makeHelloEndpoint())
```

### 4. Serve ke Transport Server

**Via GoFiber:**
```go
import "github.com/addpur/go-polykit/transport"

app := fiber.New()
app.Get("/hello", transport.NewFiberServer(
    endpoint,
    func(c *fiber.Ctx) (interface{}, error) { ... decode ... },
    func(c *fiber.Ctx, res interface{}) error { return c.JSON(res) },
    nil,
))
```

**Via Gorilla Mux:**
```go
r := mux.NewRouter()
r.Handle("/hello", transport.NewHTTPServer(
    endpoint,
    func(r *http.Request) (interface{}, error) { ... decode ... },
    func(w http.ResponseWriter, r *http.Request, res interface{}) error { ... encode ... },
    nil,
)).Methods(http.MethodGet)
```

**Via gRPC:**
```go
grpcHandler := transport.NewGRPCServer(
    endpoint,
    func(ctx context.Context, req interface{}) (interface{}, error) { ... },
    func(ctx context.Context, res interface{}) (interface{}, error) { ... },
)
```

### 5. Contoh Penggunaan via cURL

**Basic Auth (via GoFiber port 3000):**
```bash
curl -u admin:s3cr3t "http://localhost:3000/fiber/basic-secret?query=hello"
```

**JWT Bearer Token (via Gorilla Mux port 8080):**
```bash
curl -H "Authorization: Bearer <token>" "http://localhost:8080/mux/jwt-secret?query=hello"
```

*Lihat folder `example/server/main.go` untuk contoh integrasi Server selengkapnya.*

## 🛠 Linter (Clean Code Validation)
Proyek ini divalidasi menggunakan `golangci-lint` yang dikonfigurasi melalui `.golangci.yml`. Linter ini memvalidasi keamanan, pengecekan *error handler*, tipe variabel, dan penulisan standar Golang.

Untuk menjalankan linter:
```bash
golangci-lint run
```
