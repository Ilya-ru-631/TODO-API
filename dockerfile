FROM golang:1.26.4 AS build
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download 
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server 

FROM alpine:3.24
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -D -S -G appgroup appuser
WORKDIR /app
COPY --from=build /app/server /app/server
USER appuser
CMD ["./server"]