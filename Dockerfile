# Backend Build stage
ARG builderimage=golang:1.21.1-bullseye

FROM ${builderimage} as backend
WORKDIR /app
COPY src/backend .
# Copy the necessary proto files and directories
COPY src/proto/ ./proto
ENV GIN_MODE=release
# Install Protocol Buffers Compiler
RUN apt-get update && apt-get install -y protobuf-compiler
# Install the Go plugins for Protocol Buffers
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN protoc \
    --go_out=./ --go_opt=module=passgen \
    --go-grpc_out=./ --go-grpc_opt=module=passgen \
    proto/passgen.proto
RUN CGO_ENABLED=0 go build -p 4 -ldflags="-s -w" -o password-generator .

FROM ghcr.io/mindthecap/upx-container as compressor
COPY --from=backend /app/password-generator /app/password-generator
CMD ["upx", "--best", "-qq", "/app/password-generator"]

FROM ${builderimage} as gotest
WORKDIR /app
COPY src/backend .
RUN go test
RUN cd passgen && go test

# Frontend Build stage
FROM node:19.9 AS frontend
WORKDIR /app
COPY ./src/frontend/package*.json ./
RUN npm install
COPY ./src/frontend ./
# Install Protocol Buffers Compiler
RUN apt-get update && apt-get install -y protobuf-compiler
# Copy the necessary proto files and directories
COPY src/proto/ ./proto
# Run the protoc command
RUN protoc --plugin=protoc-gen-grpc-web=./node_modules/.bin/protoc-gen-grpc-web passgen.proto \
    --proto_path=./proto \
    --js_out=import_style=commonjs:./src/components \
    --grpc-web_out=import_style=typescript,mode=grpcwebtext:./src/components
RUN npm run build

# Final stage
FROM scratch
WORKDIR /app
COPY /src/backend/wordlists /app/wordlists
COPY /src/backend/values /app/values
COPY --from=compressor /app/password-generator .
COPY --from=frontend /app/dist /app
EXPOSE 8080
CMD ["./password-generator"]