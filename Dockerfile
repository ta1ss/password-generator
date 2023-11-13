# Backend Build stage
ARG builderimage=golang:1.21.1-bullseye

FROM ${builderimage} as backend-build
WORKDIR /app
COPY src/backend .
ENV GIN_MODE=release
RUN CGO_ENABLED=0 go build -p 4 -ldflags="-s -w" -o password-generator .

FROM paketobuildpacks/upx as backend
COPY --from=backend-build /app/password-generator /app/password-generator
CMD ["upx", "--ultra-brute", "-qq", "/app/password-generator"]

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
RUN npm run build

# Final stage
FROM scratch
WORKDIR /app
COPY /src/backend/wordlists /app/wordlists
COPY /src/backend/values /app/values
COPY --from=backend /app/password-generator .
COPY --from=frontend /app/dist /app
EXPOSE 8080
CMD ["./password-generator"]