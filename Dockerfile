FROM centos:7.2.1511

RUN mkdir -p /opt/ftserver
ADD ftserver /opt/ftserver/ftserver
RUN chmod +x /opt/ftserver/ftserver

RUN mkdir -p /opt/ftserver/log
RUN mkdir -p /opt/ftserver/conf
RUN mkdir -p /opt/ftserver/.cert

WORKDIR /opt/ftserver
RUN cd /opt/ftserver

CMD [ "./ftserver" ]