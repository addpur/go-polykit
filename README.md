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
3. **Context Propagation & Standarisasi**: Secara otomatis mengambil token JWT dari HTTP Header, WebSocket Query URL, atau gRPC Metadata dan menyuntikkannya ke dalam Context (Propagasi). Pustaka ini juga memperkenalkan mekanisme balasan (`StandardResponse`) yang konsisten agar mudah dikonsumsi *client*.
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

// Request
type HelloReq struct { Name string }

// Endpoint Business Logic
func makeHelloEndpoint() polykit.Endpoint {
    return func(ctx context.Context, request interface{}) (interface{}, error) {
        req := request.(HelloReq)
        
        // Context Propagation: Membaca JWT ID dari middleware
        userID := ctx.Value("user_id")

        return polykit.StandardResponse{
            ResponseCode: "00",
            Message: "Success",
            Data: "Hello " + req.Name + " (UserID: " + fmt.Sprintf("%v", userID) + ")",
        }, nil
    }
}
```

### 2. Memasang Middleware
Bungkus Endpoint dengan *chain middleware* yang telah disediakan, contohnya: JWT Validator.

```go
endpoint := makeHelloEndpoint()
endpoint = polykit.LoggingMiddleware(log.Default())(endpoint)
endpoint = polykit.JWTAuthMiddleware("secret_key")(endpoint)
```

### 3. Serve ke Transport Server
Ubah Endpoint menjadi handler framework yang Anda sukai.

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

**Via gRPC:**
```go
grpcHandler := transport.NewGRPCServer(
    endpoint,
    func(ctx context.Context, req interface{}) (interface{}, error) { ... },
    func(ctx context.Context, res interface{}) (interface{}, error) { ... },
)
```

*Lihat folder `example/main.go` dan `example/client_main.go` untuk contoh integrasi Server dan Client selengkapnya.*

## 🛠 Linter (Clean Code Validation)
Proyek ini divalidasi menggunakan `golangci-lint` yang dikonfigurasi melalui `.golangci.yml`. Linter ini memvalidasi keamanan, pengecekan *error handler*, tipe variabel, dan penulisan standar Golang.

Untuk menjalankan linter:
```bash
# Pastikan golangci-lint sudah terinstall
golangci-lint run
```
