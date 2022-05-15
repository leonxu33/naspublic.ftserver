FROM centos:7.2.1511

RUN mkdir -p /opt/ftserver
ADD ftserver /opt/ftserver/ftserver

WORKDIR /opt/ftserver
RUN cd /opt/ftserver

CMD [ "./ftserver" ]