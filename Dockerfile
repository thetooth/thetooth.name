FROM progrium/busybox
MAINTAINER Jeffrey Jenner <thetooth@ameoto.com>

COPY ./static /opt/thetooth.name/static
COPY ./template.html /opt/thetooth.name/template.html
COPY ./server /opt/thetooth.name/server
RUN chmod 755 /opt/thetooth.name/server
WORKDIR "/opt/thetooth.name"

EXPOSE 9000

CMD ["./server"]
