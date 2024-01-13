# builder stage
FROM golang:1.21.5-alpine3.19 as builder

WORKDIR /app

# these layers can be reused because of caching mechanism
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o email-service .
RUN chmod +x email-service

# Final stage
FROM alpine:3.15

WORKDIR /app

COPY --from=builder /app/email-service .
COPY --from=builder /app/.env .


CMD [ "./email-service" ]