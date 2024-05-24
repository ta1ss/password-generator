# Backend Build stage
ARG builderimage=golang:1.21.1-bullseye
ARG frontendbuildimage=node:22.1.0
# ARG swaggerimage=ghcr.io/swaggo/swag:latest
ARG swaggerimage=ghcr.io/tnaroska/swag:vx.x.x-test

FROM ${swaggerimage} as swagger
WORKDIR /app
COPY src/backend /app
RUN ["swag", "init"]

FROM ${builderimage} as backend
WORKDIR /app
COPY src/backend .
COPY --from=swagger /app/docs /app/docs
ENV GIN_MODE=release
RUN CGO_ENABLED=0 go build -p 4 -ldflags="-s -w" -o password-generator .

FROM ghcr.io/mindthecap/upx-container as compressor
COPY --from=backend /app/password-generator /app/password-generator
RUN ["upx", "--best", "-qq", "/app/password-generator"]

FROM ${builderimage} as gotest
WORKDIR /app
COPY src/backend .
RUN go test ./...

# Frontend npm install
FROM ${frontendbuildimage} as npminstall
WORKDIR /app
COPY src/frontend .
RUN npm install

# Frontend Lint stage
FROM ${frontendbuildimage} as npmlint
WORKDIR /app
COPY --from=npminstall /app .
RUN npm run lint

# Frontend Build stage
FROM ${frontendbuildimage} AS frontend
WORKDIR /app
COPY --from=npminstall /app .
RUN npm run build

# Final stage
FROM scratch
WORKDIR /app
COPY --from=compressor /app/password-generator .
COPY --from=frontend /app/dist /app
COPY --from=frontend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=backend /app/wordlists /app/wordlists
COPY --from=backend /app/values /app/values
EXPOSE 8080
CMD ["./password-generator"]