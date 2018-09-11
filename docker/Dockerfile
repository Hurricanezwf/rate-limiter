FROM alpine

ENV HOME /rate-limiter
ENV APPName rate-limiter


RUN mkdir -p $HOME/bin && mkdir -p $HOME/conf && mkdir -p $HOME/logs && mkdir -p $HOME/meta


COPY entrypoint.sh $HOME/
COPY $APPName $HOME/bin/$APPName
RUN chmod +x $HOME/bin/$APPName


ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone


EXPOSE 20000
VOLUME $HOME/logs
WORKDIR $HOME

CMD ["/bin/sh", "entrypoint.sh"]
