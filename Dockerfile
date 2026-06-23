FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o eks-monitor .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/eks-monitor /eks-monitor
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/eks-monitor"]
