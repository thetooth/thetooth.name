FROM progrium/busybox
MAINTAINER Jeffrey Jenner <thetooth@ameoto.com>

COPY ./static /opt/server/static
COPY ./template.html /opt/server/template.html
COPY ./server /opt/server/server
RUN chmod 755 /opt/server/server
WORKDIR "/opt/server"

CMD ["./server"]
