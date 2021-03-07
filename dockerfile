FROM golang:latest
RUN mkdir /app
RUN mkdir /app/mnt
ADD . /app/
WORKDIR /app
ENV TZ=Asia/Yekaterinburg
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN go get github.com/go-telegram-bot-api/telegram-bot-api
RUN go get github.com/fatih/structs
RUN go get github.com/mitchellh/mapstructure
RUN go build -o main .
CMD ["/app/main"]
