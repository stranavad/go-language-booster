FROM golang:1.21-alpine3.19
WORKDIR /usr/src/app
COPY ./ ./
RUN go mod download
RUN go build -o ./bin/language-booster ./
EXPOSE 8080
CMD [ "./bin/language-booster" ]
